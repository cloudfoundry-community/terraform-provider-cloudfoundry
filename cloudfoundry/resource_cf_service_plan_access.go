package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
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
			"plan": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"org": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"public"},
			},
			"public": &schema.Schema{
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"org"},
			},
		},
	}
}

func resourceServicePlanAccessCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	plan := d.Get("plan").(string)
	public, hasPublic := d.GetOkExists("public")
	org, hasOrg := d.GetOk("org")

	var id string
	sm := session.ServiceManager()

	if hasOrg {
		if id, err = sm.CreateServicePlanAccess(plan, org.(string)); err != nil {
			return
		}
	} else {
		state := false
		if hasPublic {
			state = public.(bool)
		}
		if err = sm.UpdateServicePlanVisibility(plan, state); err != nil {
			return
		}
		id = plan
	}

	d.SetId(id)
	return nil
}

func resourceServicePlanAccessRead(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	_, hasOrg := d.GetOk("org")

	sm := session.ServiceManager()

	if hasOrg {
		var plan, org string
		if plan, org, err = sm.ReadServicePlanAccess(d.Id()); err != nil {
			d.SetId("")
			return
		}
		d.Set("plan", plan)
		d.Set("org", org)
	} else {
		var plan cfapi.CCServicePlan
		if plan, err = sm.ReadServicePlan(d.Id()); err != nil {
			d.SetId("")
			return
		}
		d.Set("plan", d.Id())
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
