package cloudfoundry

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	resources "code.cloudfoundry.org/cli/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSegment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSegmentCreate,
		ReadContext:   resourceSegmentRead,
		UpdateContext: resourceSegmentUpdate,
		DeleteContext: resourceSegmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: ImportReadContext(resourceSegmentRead),
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
		},
	}
}

func resourceSegmentEntitlement() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSegmentEntitlementCreate,
		ReadContext:   resourceSegmentEntitlementRead,
		UpdateContext: resourceSegmentEntitlementUpdate,
		DeleteContext: resourceSegmentEntitlementDelete,
		// Importer: &schema.ResourceImporter{
		// 	State: ImportStatePassthrough,
		// },

		Schema: map[string]*schema.Schema{
			"segment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"orgs": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
				MinItems: 1,
				Required: true,
			},
			"default": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     "false",
				Description: "Set this isolation segment defined as default segment for those organizations. Default to false.",
			},
		},
	}
}

func resourceSegmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	name := d.Get("name").(string)

	sm := session.ClientV3
	seg, _, err := sm.CreateIsolationSegment(resources.IsolationSegment{
		Name: name,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(seg.GUID)
	err = metadataCreate(segmentMetadata, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSegmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(metadataUpdate(segmentMetadata, d, meta))
}

func resourceSegmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	sm := session.ClientV3
	seg, _, err := sm.GetIsolationSegment(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	d.Set("name", seg.Name)

	err = metadataRead(segmentMetadata, d, meta, false)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSegmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	sm := session.ClientV3
	_, err := sm.DeleteIsolationSegment(d.Id())
	return diag.FromErr(err)
}

func resourceSegmentEntitlementCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	segmentId := d.Get("segment").(string)
	tfOrgs := d.Get("orgs").(*schema.Set).List()
	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	orgs := make([]string, len(tfOrgs))
	for i := 0; i < len(tfOrgs); i++ {
		orgs[i] = fmt.Sprint(tfOrgs[i])
	}

	sm := session.ClientV3
	_, _, err = sm.EntitleIsolationSegmentToOrganizations(segmentId, orgs)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	if d.Get("default").(bool) {
		for _, org := range orgs {
			_, _, err := sm.UpdateOrganizationDefaultIsolationSegmentRelationship(org, segmentId)
			if err != nil {
				d.SetId("")
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func resourceSegmentEntitlementUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	sm := session.ClientV3
	if !d.HasChange("orgs") {
		return nil
	}
	segmentId := d.Get("segment").(string)
	oldTfOrgs, newTfOrgs := d.GetChange("orgs")
	toDelete, toAdd := getListChanges(oldTfOrgs, newTfOrgs)

	for _, orgId := range toDelete {
		_, err := sm.DeleteIsolationSegmentOrganization(segmentId, orgId)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if len(toAdd) > 0 {
		_, _, err := sm.EntitleIsolationSegmentToOrganizations(segmentId, toAdd)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if !d.HasChange("default") {
		return nil
	}
	tfOrgs := newTfOrgs.(*schema.Set).List()
	_, defaultChange := d.GetChange("default")
	if !defaultChange.(bool) {
		segmentId = ""
	}
	for _, tfOrg := range tfOrgs {
		_, _, err := sm.UpdateOrganizationDefaultIsolationSegmentRelationship(tfOrg.(string), segmentId)
		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceSegmentEntitlementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	sm := session.ClientV3
	orgs, _, err := sm.GetIsolationSegmentOrganizations(d.Get("segment").(string))
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	tfOrgs := d.Get("orgs").(*schema.Set).List()
	finalTfOrgs := make([]interface{}, 0)
	for _, org := range orgs {
		inside := false
		for _, tfOrg := range tfOrgs {
			if tfOrg.(string) == org.GUID {
				inside = true
				break
			}
		}
		if inside {
			finalTfOrgs = append(finalTfOrgs, org.GUID)
		}
	}
	d.Set("orgs", schema.NewSet(resourceStringHash, finalTfOrgs))
	return nil
}

func resourceSegmentEntitlementDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	sm := session.ClientV3
	tfOrgs := d.Get("orgs").(*schema.Set).List()
	for _, org := range tfOrgs {
		if d.Get("default").(bool) {
			_, _, err := sm.UpdateOrganizationDefaultIsolationSegmentRelationship(org.(string), "")
			if err != nil {
				return diag.FromErr(err)
			}
		}
		_, err := sm.DeleteIsolationSegmentOrganization(d.Get("segment").(string), fmt.Sprint(org))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
