package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	uuid "github.com/satori/go.uuid"
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
		nameOrId          string
		space             string
		serviceInstanceV3 resources.ServiceInstance
		query             ccv3.Query
	)

	nameOrId = d.Get("name_or_id").(string)
	space = d.Get("space").(string)
	isUUID := uuid.FromStringOrNil(nameOrId)
	if uuid.Equal(isUUID, uuid.Nil) {
		serviceInstance, _, _, err := session.ClientV3.GetServiceInstanceByNameAndSpace(nameOrId, space)

		if err != nil {
			return diag.FromErr(err)
		}

		serviceInstanceV3 = serviceInstance

	} else {
		query = ccv3.Query{
			Key:    ccv3.GUIDFilter,
			Values: []string{nameOrId},
		}

		serviceInstances, _, _, err := session.ClientV3.GetServiceInstances(query)

		if err != nil {
			return diag.FromErr(err)
		}

		if len(serviceInstances) == 0 {
			return diag.FromErr(ccerror.ServiceInstanceNotFoundError{
				Name:      nameOrId,
				SpaceGUID: space,
			})
		}

		serviceInstanceV3 = serviceInstances[0]
	}

	d.SetId(serviceInstanceV3.GUID)
	d.Set("name", serviceInstanceV3.Name)
	d.Set("service_plan_id", serviceInstanceV3.ServicePlanGUID)
	if serviceInstanceV3.Tags.IsSet {
		tags := make([]interface{}, len(serviceInstanceV3.Tags.Value))
		for i, v := range serviceInstanceV3.Tags.Value {
			tags[i] = v
		}
		d.Set("tags", tags)
	} else {
		d.Set("tags", nil)
	}

	return nil
}
