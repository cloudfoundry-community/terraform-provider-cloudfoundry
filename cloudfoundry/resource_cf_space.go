package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceSpace() *schema.Resource {

	return &schema.Resource{

		Create: resourceSpaceCreate,
		Read:   resourceSpaceRead,
		Update: resourceSpaceUpdate,
		Delete: resourceSpaceDelete,

		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"quota": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"allow_ssh": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"isolation_segment": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"asgs": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"staging_asgs": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"managers": &schema.Schema{
				Deprecated: "Use resource cloudfoundry_space_users instead",
				Type:       schema.TypeSet,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			"developers": &schema.Schema{
				Deprecated: "Use resource cloudfoundry_space_users instead",
				Type:       schema.TypeSet,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			"auditors": &schema.Schema{
				Deprecated: "Use resource cloudfoundry_space_users instead",
				Type:       schema.TypeSet,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
		},
	}
}

func resourceSpaceCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	org := d.Get("org").(string)
	quota := d.Get("quota").(string)
	allowSSH := d.Get("allow_ssh").(bool)

	sm := session.ClientV2
	space, _, err := sm.CreateSpaceFromObject(ccv2.Space{
		Name:                     name,
		OrganizationGUID:         org,
		AllowSSH:                 allowSSH,
		SpaceQuotaDefinitionGUID: quota,
	})
	if err != nil {
		return err
	}
	d.SetId(space.GUID)
	err = resourceSpaceUpdate(d, NewResourceMeta{meta})
	if err != nil {
		return err
	}

	return nil
}

func resourceSpaceRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	id := d.Id()
	sm := session.ClientV2

	space, _, err := sm.GetSpace(id)
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	d.Set("name", space.Name)
	d.Set("org", space.OrganizationGUID)
	d.Set("quota", space.SpaceQuotaDefinitionGUID)
	d.Set("allow_ssh", space.AllowSSH)

	for t, r := range typeToSpaceRoleMap {
		users, _, err := sm.GetSpaceUsersByRole(r, id)
		if err != nil {
			return err
		}
		tfUsers := d.Get(t).(*schema.Set).List()
		if !IsImportState(d) {
			finalUsers := intersectSlices(tfUsers, users, func(source, item interface{}) bool {
				return source.(string) == item.(ccv2.User).GUID
			})
			d.Set(t, schema.NewSet(resourceStringHash, finalUsers))
		} else {
			d.Set(t, schema.NewSet(resourceStringHash, objectsToIds(users, func(object interface{}) string {
				return object.(ccv2.User).GUID
			})))
		}
	}

	runningAsgs, _, err := session.ClientV2.GetSpaceSecurityGroups(d.Id())
	if err != nil {
		return nil
	}
	if !IsImportState(d) {
		finalRunningAsg := intersectSlices(d.Get("asgs").(*schema.Set).List(), runningAsgs, func(source, item interface{}) bool {
			return source.(string) == item.(ccv2.SecurityGroup).GUID
		})
		d.Set("asgs", schema.NewSet(resourceStringHash, finalRunningAsg))
	} else {
		d.Set("asgs", schema.NewSet(resourceStringHash, objectsToIds(runningAsgs, func(object interface{}) string {
			return object.(ccv2.SecurityGroup).GUID
		})))
	}

	stagingAsgs, _, err := session.ClientV2.GetSpaceStagingSecurityGroups(d.Id())
	if err != nil {
		return nil
	}
	if !IsImportState(d) {
		finalStagingAsg := intersectSlices(d.Get("staging_asgs").(*schema.Set).List(), stagingAsgs, func(source, item interface{}) bool {
			return source.(string) == item.(ccv2.SecurityGroup).GUID
		})
		d.Set("staging_asgs", schema.NewSet(resourceStringHash, finalStagingAsg))
	} else {
		d.Set("staging_asgs", schema.NewSet(resourceStringHash, objectsToIds(stagingAsgs, func(object interface{}) string {
			return object.(ccv2.SecurityGroup).GUID
		})))
	}

	segment, _, err := session.ClientV3.GetSpaceIsolationSegment(d.Id())
	if err != nil {
		return err
	}
	d.Set("isolation_segment", segment.GUID)
	return nil
}

func resourceSpaceUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	var (
		newResource bool
		session     *managers.Session
	)

	if m, ok := meta.(NewResourceMeta); ok {
		session = m.meta.(*managers.Session)
		newResource = true
	} else {
		session = meta.(*managers.Session)
		newResource = false
	}

	spaceID := d.Id()
	orgID := d.Get("org").(string)
	if !newResource {
		_, _, err := session.ClientV2.UpdateSpace(ccv2.Space{
			GUID:             spaceID,
			Name:             d.Get("name").(string),
			OrganizationGUID: orgID,
			AllowSSH:         d.Get("allow_ssh").(bool),
		})
		if err != nil {
			return err
		}
		if d.HasChange("quota") {
			_, err := session.ClientV2.SetSpaceQuota(spaceID, d.Get("quota").(string))
			if err != nil {
				return err
			}
		}
	}

	removeAsgs, addAsgs := getListChanges(d.GetChange("asgs"))
	for _, asgID := range removeAsgs {
		_, err = session.ClientV2.DeleteSecurityGroupSpace(asgID, spaceID)
		if err != nil {
			return err
		}
	}
	for _, asgID := range addAsgs {
		_, err = session.ClientV2.UpdateSecurityGroupSpace(asgID, spaceID)
		if err != nil {
			return err
		}
	}

	removeStagingAsgs, addStagingAsgs := getListChanges(d.GetChange("staging_asgs"))
	for _, asgID := range removeStagingAsgs {
		_, err = session.ClientV2.DeleteSecurityGroupStagingSpace(asgID, spaceID)
		if err != nil {
			return err
		}
	}
	for _, asgID := range addStagingAsgs {
		_, err = session.ClientV2.UpdateSecurityGroupStagingSpace(asgID, spaceID)
		if err != nil {
			return err
		}
	}

	for t, r := range typeToSpaceRoleMap {
		remove, add := getListChanges(d.GetChange(t))
		for _, uid := range remove {
			_, err = session.ClientV2.DeleteSpaceUserByRole(r, spaceID, uid)
			if err != nil {
				return err
			}
		}
		for _, uid := range add {
			err = addOrNothingUserInOrgBySpace(session.ClientV2, orgID, uid)
			if err != nil {
				return err
			}
			_, err = session.ClientV2.UpdateSpaceUserByRole(r, spaceID, uid)
			if err != nil {
				return err
			}
		}
	}

	segID := d.Get("isolation_segment").(string)
	if segID != "" && newResource {
		_, _, err := session.ClientV3.UpdateSpaceIsolationSegmentRelationship(spaceID, segID)
		if err != nil {
			return err
		}
	}

	if !newResource && d.HasChange("isolation_segment") {
		_, _, err := session.ClientV3.UpdateSpaceIsolationSegmentRelationship(spaceID, segID)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceSpaceDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	j, _, err := session.ClientV2.DeleteSpace(d.Id())
	_, err = session.ClientV2.PollJob(j)
	if err != nil {
		return err
	}
	return err
}
