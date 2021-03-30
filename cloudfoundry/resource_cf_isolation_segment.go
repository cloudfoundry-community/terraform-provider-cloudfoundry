package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSegment() *schema.Resource {
	return &schema.Resource{
		Create: resourceSegmentCreate,
		Read:   resourceSegmentRead,
		Update: resourceSegmentUpdate,
		Delete: resourceSegmentDelete,
		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceSegmentRead),
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
		Create: resourceSegmentEntitlementCreate,
		Read:   resourceSegmentEntitlementRead,
		Update: resourceSegmentEntitlementUpdate,
		Delete: resourceSegmentEntitlementDelete,
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

func resourceSegmentCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	name := d.Get("name").(string)

	sm := session.ClientV3
	seg, _, err := sm.CreateIsolationSegment(ccv3.IsolationSegment{
		Name: name,
	})
	if err != nil {
		return err
	}
	d.SetId(seg.GUID)
	err = metadataCreate(segmentMetadata, d, meta)
	if err != nil {
		return err
	}
	return nil
}

func resourceSegmentUpdate(d *schema.ResourceData, meta interface{}) error {
	return metadataUpdate(segmentMetadata, d, meta)
}

func resourceSegmentRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	sm := session.ClientV3
	seg, _, err := sm.GetIsolationSegment(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	d.Set("name", seg.Name)

	err = metadataRead(segmentMetadata, d, meta, false)
	if err != nil {
		return err
	}
	return nil
}

func resourceSegmentDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	sm := session.ClientV3
	_, err := sm.DeleteIsolationSegment(d.Id())
	return err
}

func resourceSegmentEntitlementCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	segmentId := d.Get("segment").(string)
	tfOrgs := d.Get("orgs").(*schema.Set).List()
	id, err := uuid.GenerateUUID()
	if err != nil {
		return err
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
		return err
	}
	if d.Get("default").(bool) {
		for _, org := range orgs {
			_, _, err := sm.UpdateOrganizationDefaultIsolationSegmentRelationship(org, segmentId)
			if err != nil {
				d.SetId("")
				return err
			}
		}
	}
	return nil
}

func resourceSegmentEntitlementUpdate(d *schema.ResourceData, meta interface{}) error {
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
			return err
		}
	}
	if len(toAdd) > 0 {
		_, _, err := sm.EntitleIsolationSegmentToOrganizations(segmentId, toAdd)
		if err != nil {
			return err
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
			return err
		}
	}
	return nil
}

func resourceSegmentEntitlementRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	sm := session.ClientV3
	orgs, _, err := sm.GetIsolationSegmentOrganizations(d.Get("segment").(string))
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
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

func resourceSegmentEntitlementDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	sm := session.ClientV3
	tfOrgs := d.Get("orgs").(*schema.Set).List()
	for _, org := range tfOrgs {
		if d.Get("default").(bool) {
			_, _, err := sm.UpdateOrganizationDefaultIsolationSegmentRelationship(org.(string), "")
			if err != nil {
				return err
			}
		}
		_, err := sm.DeleteIsolationSegmentOrganization(d.Get("segment").(string), fmt.Sprint(org))
		if err != nil {
			return err
		}
	}

	return nil
}
