package cloudfoundry

import (
	"context"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServiceOfferingV3() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceServiceOfferingV3Read,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"service_broker_guid": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"service_broker_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_plans": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceOfferingV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	space := d.Get("space").(string)
	serviceBrokerGUID := d.Get("service_broker_guid").(string)

	filters := []ccv3.Query{}

	// Add required filter
	filters = append(filters, ccv3.Query{
		Key:    ccv3.NameFilter,
		Values: []string{name},
	})

	if serviceBrokerGUID != "" {
		filters = append(filters, ccv3.Query{
			Key:    ccv3.ServiceBrokerGUIDsFilter,
			Values: []string{serviceBrokerGUID},
		})
	}

	if space != "" {
		filters = append(filters, ccv3.Query{
			Key:    ccv3.SpaceGUIDFilter,
			Values: []string{space},
		})
	}

	tflog.Debug(ctx, fmt.Sprintf("%+v", filters))

	services, _, err := session.ClientV3.GetServiceOfferings(filters...)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(services) == 0 {
		return diag.FromErr(NotFound)
	}
	service := services[0]

	d.SetId(service.GUID)
	if serviceBrokerGUID == "" {
		d.Set("service_broker_name", service.ServiceBrokerName)
	}

	// Get service plans
	servicePlans, _, err := session.ClientV3.GetServicePlans(ccv3.Query{
		// Constant not defined by cli ccv3
		Key:    "service_offering_guids",
		Values: []string{service.GUID},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	servicePlansTf := make(map[string]interface{})
	for _, sp := range servicePlans {
		servicePlansTf[strings.Replace(sp.Name, ".", "_", -1)] = sp.GUID
	}
	d.Set("service_plans", servicePlansTf)

	return nil
}
