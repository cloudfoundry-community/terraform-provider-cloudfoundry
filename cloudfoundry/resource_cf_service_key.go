package cloudfoundry

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceServiceKey() *schema.Resource {

	return &schema.Resource{

		Create: resourceServiceKeyCreate,
		Read:   resourceServiceKeyRead,
		Delete: resourceServiceKeyDelete,

		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_instance": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"params": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"credentials": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceServiceKeyCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	name := d.Get("name").(string)
	serviceInstance := d.Get("service_instance").(string)
	params := d.Get("params").(map[string]interface{})

	sm := session.ServiceManager()
	var serviceKey cfapi.CCServiceKey

	if serviceKey, err = sm.CreateServiceKey(name, serviceInstance, params); err != nil {
		return err
	}
	session.Log.DebugMessage("Created Service Key: %# v", serviceKey)

	d.Set("credentials", normalizeMap(serviceKey.Credentials, make(map[string]interface{}), "", "_"))
	d.SetId(serviceKey.ID)
	return nil
}

func resourceServiceKeyRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("Reading Service Key with ID: %s", d.Id())

	sm := session.ServiceManager()
	var serviceKey cfapi.CCServiceKey

	if serviceKey, err = sm.ReadServiceKey(d.Id()); err != nil {
		if strings.Contains(err.Error(), "status code: 404") {
			d.SetId("")
			err = nil
		}
		return err
	}
	d.Set("name", serviceKey.Name)
	d.Set("service_instance", serviceKey.ServiceGUID)
	d.Set("credentials", normalizeMap(serviceKey.Credentials, make(map[string]interface{}), "", "_"))

	session.Log.DebugMessage("Read Service Instance : %# v", serviceKey)
	return nil
}

func resourceServiceKeyDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("Reading Service Key with ID: %s", d.Id())

	err = session.ServiceManager().DeleteServiceKey(d.Id())
	return err
}
