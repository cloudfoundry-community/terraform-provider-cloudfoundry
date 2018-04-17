package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func dataServiceKey() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceServiceKeyRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"service_instance": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"credentials": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceKeyRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("Reading Service Key with name: %s", d.Get("name").(string))

	sm := session.ServiceManager()
	var serviceKey cfapi.CCServiceKey

	serviceKey, err = sm.FindServiceKey(d.Get("name").(string), d.Get("service_instance").(string))
	if err != nil {
		return
	}
	d.SetId(serviceKey.ID)
	d.Set("name", serviceKey.Name)
	d.Set("service_instance", serviceKey.ServiceGUID)
	d.Set("credentials", normalizeMap(serviceKey.Credentials, make(map[string]interface{}), "", "_"))

	session.Log.DebugMessage("Read Service Instance : %# v", serviceKey)
	return
}
