package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServicePlanAccessImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	session := meta.(*managers.Session)
	if session == nil {
		return []*schema.ResourceData{}, fmt.Errorf("client is nil")
	}

	spV, _, err := session.ClientV2.GetServicePlanVisibility(d.Id())
	if err == nil {
		d.Set("plan", spV.ServicePlanGUID)
		d.Set("org", spV.OrganizationGUID)
		return ImportRead(resourceServicePlanAccessRead)(d, meta)
	}

	plan, _, err := session.ClientV2.GetServicePlan(d.Id())
	if err == nil {
		d.Set("plan", d.Id())
		d.Set("public", plan.Public)
		return ImportRead(resourceServicePlanAccessRead)(d, meta)
	}

	return []*schema.ResourceData{}, fmt.Errorf("unable to find service_plan_visibilities nor service plan for id '%s'", d.Id())
}
