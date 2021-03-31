package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/satori/go.uuid"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func dataSourceServiceInstance() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceServiceInstanceRead,

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

func dataSourceServiceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	var (
		nameOrId        string
		space           string
		serviceInstance ccv2.ServiceInstance
	)

	nameOrId = d.Get("name_or_id").(string)
	space = d.Get("space").(string)
	isUUID := uuid.FromStringOrNil(nameOrId)
	if uuid.Equal(isUUID, uuid.Nil) {
		serviceInstances, _, err := session.ClientV2.GetServiceInstances(ccv2.FilterByName(nameOrId), ccv2.FilterEqual(constant.SpaceGUIDFilter, space))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(serviceInstances) == 0 {
			return diag.FromErr(NotFound)
		}
		serviceInstance = serviceInstances[0]
	} else {
		var err error
		serviceInstance, _, err = session.ClientV2.GetServiceInstance(nameOrId)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(serviceInstance.GUID)
	d.Set("name", serviceInstance.Name)
	d.Set("service_plan_id", serviceInstance.ServicePlanGUID)
	d.Set("tags", serviceInstance.Tags)

	return nil
}
