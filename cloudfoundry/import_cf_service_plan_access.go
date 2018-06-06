package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceServicePlanAccessImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return []*schema.ResourceData{}, fmt.Errorf("client is nil")
	}
	sm := session.ServiceManager()

	planID, orgID, err := sm.ReadServicePlanAccess(d.Id())
	if err == nil {
		d.Set("plan", planID)
		d.Set("org", orgID)
		return schema.ImportStatePassthrough(d, meta)
	}

	plan, err := sm.ReadServicePlan(d.Id())
	if err == nil {
		d.Set("plan", d.Id())
		d.Set("public", plan.Public)
		return schema.ImportStatePassthrough(d, meta)
	}

	return []*schema.ResourceData{}, fmt.Errorf("unable to find service_plan_visibilities nor service plan for id '%s'", d.Id())
}
