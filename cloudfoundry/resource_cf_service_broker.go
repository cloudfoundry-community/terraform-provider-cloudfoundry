package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceServiceBroker() *schema.Resource {

	return &schema.Resource{

		Create: resourceServiceBrokerCreate,
		Read:   resourceServiceBrokerRead,
		Update: resourceServiceBrokerUpdate,
		Delete: resourceServiceBrokerDelete,

		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"url": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"service_plans": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceServiceBrokerCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	_, name, url, username, password, space := getSchemaAttributes(d)
	sb, _, err := session.ClientV2.CreateServiceBroker(name, url, username, password, space)
	if err != nil {
		return err
	}
	if err = readServiceDetail(sb.GUID, session, d); err != nil {
		return err
	}
	d.SetId(sb.GUID)
	return nil
}

func resourceServiceBrokerRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	sb, _, err := session.ClientV2.GetServiceBroker(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	err = readServiceDetail(d.Id(), session, d)
	if err != nil {
		d.SetId("")
		return err
	}

	d.Set("name", sb.Name)
	d.Set("url", sb.BrokerURL)
	d.Set("username", sb.AuthUsername)
	d.Set("space", sb.SpaceGUID)

	return err
}

func resourceServiceBrokerUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	id, name, url, username, password, space := getSchemaAttributes(d)

	_, _, err := session.ClientV2.UpdateServiceBroker(ccv2.ServiceBroker{
		GUID:         d.Id(),
		AuthUsername: username,
		AuthPassword: password,
		BrokerURL:    url,
		SpaceGUID:    space,
		Name:         name,
	})
	if err != nil {
		d.SetId("")
		return err
	}
	if err = readServiceDetail(id, session, d); err != nil {
		d.SetId("")
		return err
	}

	return err
}

func resourceServiceBrokerDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	if !session.PurgeWhenDelete {
		_, err := session.ClientV2.DeleteServiceBroker(d.Id())
		return err
	}

	svcs, _, err := session.ClientV2.GetServices(ccv2.FilterEqual(constant.ServiceBrokerGUIDFilter, d.Id()))
	if err != nil {
		return err
	}
	for _, svc := range svcs {
		sis, _, err := session.ClientV2.GetServiceInstances(ccv2.FilterEqual(constant.ServiceGUIDFilter, svc.GUID))
		if err != nil {
			return err
		}
		for _, si := range sis {
			_, _, err := session.ClientV2.DeleteServiceInstance(si.GUID, true, true)
			if err != nil {
				return err
			}
		}
	}
	_, err = session.ClientV2.DeleteServiceBroker(d.Id())
	return err
}

func getSchemaAttributes(d *schema.ResourceData) (id, name, url, username, password, space string) {

	id = d.Id()
	name = d.Get("name").(string)
	url = d.Get("url").(string)
	username = d.Get("username").(string)
	password = d.Get("password").(string)

	if v, ok := d.GetOk("space"); ok {
		space = v.(string)
	}
	return id, name, url, username, password, space
}

func readServiceDetail(id string, session *managers.Session, d *schema.ResourceData) error {
	services, _, err := session.ClientV2.GetServices(ccv2.FilterEqual(constant.ServiceBrokerGUIDFilter, id))
	if err != nil {
		return err
	}

	servicePlansTf := make(map[string]interface{})
	for _, s := range services {
		servicePlans, _, err := session.ClientV2.GetServicePlans(ccv2.FilterEqual(constant.ServiceGUIDFilter, s.GUID))
		if err != nil {
			return err
		}
		for _, sp := range servicePlans {
			servicePlansTf[s.Label+"/"+sp.Name] = sp.GUID
		}
	}
	d.Set("service_plans", servicePlansTf)

	return err
}
