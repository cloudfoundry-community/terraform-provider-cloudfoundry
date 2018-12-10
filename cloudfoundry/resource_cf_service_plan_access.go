package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

func resourceServicePlanAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourceServicePlanAccessCreate,
		Read:   resourceServicePlanAccessRead,
		Delete: resourceServicePlanAccessDelete,
		Importer: &schema.ResourceImporter{
			State: resourceServicePlanAccessImport,
		},
		Schema: map[string]*schema.Schema{
			"plan_id": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"org_id": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"public"},
			},
			"public": &schema.Schema{
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"org_id"},
			},
		},
	}
}

func resourceServicePlanAccessCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	planID := d.Get("plan_id").(string)
	public, hasPublic := d.GetOkExists("public")
	orgID, hasOrg := d.GetOk("org_id")

	var id string
	sm := session.ServiceManager()

	if hasOrg {
		if id, err = sm.CreateServicePlanAccess(planID, orgID.(string)); err != nil {
			return
		}
	} else {
		state := false
		if hasPublic {
			state = public.(bool)
		}
		if err = sm.UpdateServicePlanVisibility(planID, state); err != nil {
			return
		}
		id = planID
	}

	d.SetId(id)
	return nil
}

func resourceServicePlanAccessRead(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	_, hasOrg := d.GetOk("org_id")

	sm := session.ServiceManager()

	if hasOrg {
		var planID, orgID string
		if planID, orgID, err = sm.ReadServicePlanAccess(d.Id()); err != nil {
			d.SetId("")
			return
		}
		d.Set("plan_id", planID)
		d.Set("org_id", orgID)
	} else {
		var plan cfapi.CCServicePlan
		if plan, err = sm.ReadServicePlan(d.Id()); err != nil {
			d.SetId("")
			return
		}
		d.Set("plan_id", d.Id())
		d.Set("public", plan.Public)
	}

	return nil
}

func resourceServicePlanAccessDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	_, hasOrg := d.GetOk("org")
	if hasOrg {
		sm := session.ServiceManager()
		if err = sm.DeleteServicePlanAccess(d.Id()); err != nil {
			return err
		}
	}
	return nil
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
