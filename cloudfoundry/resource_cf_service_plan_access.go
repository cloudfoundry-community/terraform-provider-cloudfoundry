package cloudfoundry

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServicePlanAccess() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServicePlanAccessCreate,
		ReadContext:   resourceServicePlanAccessRead,
		DeleteContext: resourceServicePlanAccessDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServicePlanAccessImport,
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

func resourceServicePlanAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	plan := d.Get("plan").(string)
	public, hasPublic := d.GetOkExists("public")
	org, hasOrg := d.GetOk("org")

	var id string
	if hasOrg {
		spV, _, err := session.ClientV2.CreateServicePlanVisibility(plan, org.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		id = spV.GUID
	} else {
		state := false
		if hasPublic {
			state = public.(bool)
		}
		_, err := session.ClientV2.UpdateServicePlan(plan, state)
		if err != nil {
			return diag.FromErr(err)
		}
		id = plan
	}

	d.SetId(id)
	return nil
}

func resourceServicePlanAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	_, hasOrg := d.GetOk("org")

	if hasOrg {
		spV, _, err := session.ClientV2.GetServicePlanVisibility(d.Id())
		if err != nil {
			if IsErrNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		d.Set("plan", spV.ServicePlanGUID)
		d.Set("org", spV.OrganizationGUID)
	} else {
		plan, _, err := session.ClientV2.GetServicePlan(d.Id())
		if err != nil {
			if IsErrNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		d.Set("plan", d.Id())
		d.Set("public", plan.Public)
	}

	return nil
}

func resourceServicePlanAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	_, hasOrg := d.GetOk("org")
	if !hasOrg {
		return nil
	}
	_, err := session.ClientV2.DeleteServicePlanVisibility(d.Id())
	return diag.FromErr(err)
}
