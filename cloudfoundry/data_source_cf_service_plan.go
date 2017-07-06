package cloudfoundry

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceServicePlan() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceServicePlanRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"service": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceServicePlanRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	sm := session.ServiceManager()

	var service = d.Get("service").(string)
	var name = d.Get("name").(string)
	var id string

	id, err = sm.FindServicePlanID(service, name)

	if err != nil {
		return
	}
	d.SetId(id)
	return
}
