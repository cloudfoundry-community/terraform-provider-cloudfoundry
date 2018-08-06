package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
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

func resourceServiceBrokerCreate(d *schema.ResourceData, meta interface{}) (err error) {

	var id string
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	_, name, url, username, password, space := getSchemaAttributes(d)

	sm := session.ServiceManager()

	if id, err = sm.CreateServiceBroker(name, url, username, password, space); err != nil {
		return err
	}
	if err = readServiceDetail(id, sm, d); err != nil {
		return err
	}
	session.Log.DebugMessage("Service detail for service broker: %s:\n%# v\n", name, d.Get("service_plans"))

	d.SetId(id)
	return nil
}

func resourceServiceBrokerRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		serviceBroker cfapi.CCServiceBroker
	)

	sm := session.ServiceManager()
	if serviceBroker, err = sm.ReadServiceBroker(d.Id()); err != nil {
		d.SetId("")
		return err
	}
	if err = readServiceDetail(d.Id(), sm, d); err != nil {
		d.SetId("")
		return err
	}

	d.Set("name", serviceBroker.Name)
	d.Set("url", serviceBroker.BrokerURL)
	d.Set("username", serviceBroker.AuthUserName)
	d.Set("space", serviceBroker.SpaceGUID)

	return err
}

func resourceServiceBrokerUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id, name, url, username, password, space := getSchemaAttributes(d)

	sm := session.ServiceManager()
	if _, err = sm.UpdateServiceBroker(id, name, url, username, password, space); err != nil {
		d.SetId("")
		return err
	}
	if err = readServiceDetail(id, sm, d); err != nil {
		d.SetId("")
		return err
	}
	session.Log.DebugMessage("Service detail for service broker: %s:\n%# v\n", name, d.Get("service_plans"))

	return err
}

func resourceServiceBrokerDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	sm := session.ServiceManager()
	err = sm.DeleteServiceBroker(d.Id())
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

func readServiceDetail(id string, sm *cfapi.ServiceManager, d *schema.ResourceData) (err error) {

	var (
		services []cfapi.CCService
	)

	if services, err = sm.ReadServiceInfo(id); err != nil {
		return err
	}

	servicePlans := make(map[string]interface{})
	for _, s := range services {
		for _, sp := range s.ServicePlans {
			servicePlans[s.Label+"/"+sp.Name] = sp.ID
		}
	}
	d.Set("service_plans", servicePlans)

	return err
}
