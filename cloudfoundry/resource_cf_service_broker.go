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
			State: ImportRead(resourceServiceBrokerRead),
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
				ForceNew: true,
			},
			"service_plans": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
			"services": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceServiceBrokerCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	sb, _, err := session.ClientV2.CreateServiceBroker(
		d.Get("name").(string),
		d.Get("username").(string),
		d.Get("password").(string),
		d.Get("url").(string),
		d.Get("space").(string),
	)
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

	_, _, err := session.ClientV2.UpdateServiceBroker(ccv2.ServiceBroker{
		GUID:         d.Id(),
		AuthUsername: d.Get("username").(string),
		AuthPassword: d.Get("password").(string),
		BrokerURL:    d.Get("url").(string),
		SpaceGUID:    d.Get("space").(string),
		Name:         d.Get("name").(string),
	})
	if err != nil {
		return err
	}

	if err = readServiceDetail(d.Id(), session, d); err != nil {
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

func readServiceDetail(id string, session *managers.Session, d *schema.ResourceData) error {
	services, _, err := session.ClientV2.GetServices(ccv2.FilterEqual(constant.ServiceBrokerGUIDFilter, id))
	if err != nil {
		return err
	}

	servicePlansTf := make(map[string]interface{})
	servicesTf := make(map[string]interface{})
	for _, s := range services {
		servicesTf[s.Label] = s.GUID
		servicePlans, _, err := session.ClientV2.GetServicePlans(ccv2.FilterEqual(constant.ServiceGUIDFilter, s.GUID))
		if err != nil {
			return err
		}
		for _, sp := range servicePlans {
			servicePlansTf[s.Label+"/"+sp.Name] = sp.GUID
		}
	}
	d.Set("service_plans", servicePlansTf)
	d.Set("services", servicesTf)

	return err
}
