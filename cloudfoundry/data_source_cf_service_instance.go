package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/satori/go.uuid"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func dataSourceServiceInstance() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceServiceInstanceRead,

		Schema: map[string]*schema.Schema{

			"name_or_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_plan_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceServiceInstanceRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	var (
		name_or_id      string
		space           string
		serviceInstance ccv2.ServiceInstance
	)

	name_or_id = d.Get("name_or_id").(string)
	space = d.Get("space").(string)
	isUUID := uuid.FromStringOrNil(name_or_id)
	if uuid.Equal(isUUID, uuid.Nil) {
		serviceInstances, _, err := session.ClientV2.GetServiceInstances(ccv2.FilterByName(name_or_id), ccv2.FilterEqual(constant.SpaceGUIDFilter, space))
		if err != nil {
			return err
		}
		if len(serviceInstances) == 0 {
			return NotFound
		}
		serviceInstance = serviceInstances[0]
	} else {
		var err error
		serviceInstance, _, err = session.ClientV2.GetServiceInstance(name_or_id)
		if err != nil {
			return err
		}
	}

	d.SetId(serviceInstance.GUID)
	d.Set("name", serviceInstance.Name)
	d.Set("service_plan_id", serviceInstance.ServicePlanGUID)
	d.Set("tags", serviceInstance.Tags)

	return nil
}
