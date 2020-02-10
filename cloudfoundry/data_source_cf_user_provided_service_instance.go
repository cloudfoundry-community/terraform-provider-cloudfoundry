package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func dataSourceUserProvidedServiceInstance() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceUserProvidedServiceInstanceRead,

		Schema: map[string]*schema.Schema{
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceUserProvidedServiceInstanceRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	var (
		name            string
		space           string
		serviceInstance ccv2.ServiceInstance
	)

	name = d.Get("name").(string)
	space = d.Get("space").(string)
	serviceInstances, _, err := session.ClientV2.GetUserProvidedServiceInstances(ccv2.FilterByName(name), ccv2.FilterEqual(constant.SpaceGUIDFilter, space))
	if err != nil {
		return err
	}
	if len(serviceInstances) == 0 {
		return NotFound
	}
	serviceInstance = serviceInstances[0]

	d.SetId(serviceInstance.GUID)
	d.Set("name", serviceInstance.Name)
	d.Set("tags", serviceInstance.Tags)

	return nil
}
