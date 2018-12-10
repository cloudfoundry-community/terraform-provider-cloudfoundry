package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

func resourceSpace() *schema.Resource {

	return &schema.Resource{

		Create: resourceSpaceCreate,
		Read:   resourceSpaceRead,
		Update: resourceSpaceUpdate,
		Delete: resourceSpaceDelete,

		Importer: &schema.ResourceImporter{
			State: resourceSpaceImport,
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"quota_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"allow_ssh": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"isolation_segment_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"asg_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"staging_asg_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"manager_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"developer_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"auditor_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
		},
	}
}

var typeToSpaceRoleMap = map[string]cfapi.SpaceRole{
	"manager_ids":   cfapi.SpaceRoleManager,
	"developer_ids": cfapi.SpaceRoleDeveloper,
	"auditor_ids":   cfapi.SpaceRoleAuditor,
}

func resourceSpaceCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		name, orgID, quotaID string
		allowSSH         bool
		asgIDs             []interface{}
	)
	name = d.Get("name").(string)
	orgID = d.Get("org_id").(string)
	if v, ok := d.GetOk("quota_id"); ok {
		quotaID = v.(string)
	}
	if v, ok := d.GetOk("asg_ids"); ok {
		asgIDs = v.(*schema.Set).List()
	}
	allowSSH = d.Get("allow_ssh").(bool)

	var id string

	sm := session.SpaceManager()
	if id, err = sm.CreateSpace(name, orgID, quotaID, allowSSH, asgIDs); err != nil {
		return err
	}
	d.SetId(id)

	err = resourceSpaceUpdate(d, NewResourceMeta{meta})
	if err != nil {
		return err
	}
	return nil
}

func resourceSpaceRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id := d.Id()
	sm := session.SpaceManager()

	var (
		space cfapi.CCSpace

		runningAsgIDs     []string
		spaceAsgIDs, asgIDs []interface{}
	)

	if space, err = sm.ReadSpace(id); err != nil {
		return err
	}
	d.Set("name", space.Name)
	d.Set("org_id", space.OrgGUID)
	d.Set("quota_id", space.QuotaGUID)
	d.Set("allow_ssh", space.AllowSSH)

	var users []interface{}
	for t, r := range typeToSpaceRoleMap {
		if users, err = sm.ListUsers(id, r); err != nil {
			return
		}
		d.Set(t, schema.NewSet(resourceStringHash, users))
	}

	if runningAsgIDs, err = session.ASGManager().Running(); err != nil {
		return err
	}
	if spaceAsgIDs, err = sm.ListASGs(id); err != nil {
		return err
	}
	for _, a := range spaceAsgIDs {
		if !isStringInList(runningAsgIDs, a.(string)) {
			asgIDs = append(asgIDs, a)
		}
	}
	d.Set("asg_ids", schema.NewSet(resourceStringHash, asgIDs))

	segmentID, err := sm.GetSpaceSegment(id)
	if err != nil {
		return err
	}
	d.Set("isolation_segment_id", segmentID)
	return nil
}

func resourceSpaceUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	var (
		newResource bool
		session     *cfapi.Session
	)

	if m, ok := meta.(NewResourceMeta); ok {
		session = m.meta.(*cfapi.Session)
		newResource = true
	} else {
		session = meta.(*cfapi.Session)
		if session == nil {
			return fmt.Errorf("client is nil")
		}
		newResource = false
	}

	spaceID := d.Id()
	orgID := d.Get("org_id").(string)

	om := session.OrgManager()
	sm := session.SpaceManager()

	if !newResource {

		var asgIDs []interface{}

		space := cfapi.CCSpace{
			ID:   spaceID,
			Name: d.Get("name").(string),

			OrgGUID: orgID,

			AllowSSH: d.Get("allow_ssh").(bool),
		}
		if v, ok := d.GetOk("quota_id"); ok {
			space.QuotaGUID = v.(string)
		}
		if v, ok := d.GetOk("asg_ids"); ok {
			asgIDs = v.(*schema.Set).List()
		}

		if err = sm.UpdateSpace(space, asgIDs); err != nil {
			return err
		}
	}

	old, new := d.GetChange("staging_asg_ids")
	remove, add := getListChanges(old, new)
	for _, asgID := range remove {
		if err = sm.RemoveStagingASG(spaceID, asgID); err != nil {
			return err
		}
	}
	for _, asgID := range add {
		if err = sm.AddStagingASG(spaceID, asgID); err != nil {
			return err
		}
	}

	usersRemoved := make(map[string]bool)
	usersAdded := make(map[string]bool)

	for t, r := range typeToSpaceRoleMap {
		old, new := d.GetChange(t)
		remove, add := getListChanges(old, new)

		for _, uid := range remove {
			session.Log.DebugMessage("Removing user '%s' from space '%s' with role '%s'.", uid, spaceID, r)
			if err = sm.RemoveUser(spaceID, uid, r); err != nil {
				return err
			}
		}
		for _, uid := range add {
			session.Log.DebugMessage("Adding user '%s' to space '%s' with role '%s'.", uid, spaceID, r)
			if err = om.AddUser(orgID, uid, cfapi.OrgRoleMember); err != nil {
				return err
			}
			if err = sm.AddUser(spaceID, uid, r); err != nil {
				return err
			}
		}

		for _, r := range remove {
			usersRemoved[r] = true
		}
		for _, r := range add {
			usersAdded[r] = true
		}
	}

	orgUsers := make(map[string]bool)
	for _, r := range []cfapi.OrgRole{
		cfapi.OrgRoleManager,
		cfapi.OrgRoleBillingManager,
		cfapi.OrgRoleAuditor} {

		var uu []interface{}
		if uu, err = om.ListUsers(orgID, r); err != nil {
			return err
		}
		for _, u := range uu {
			orgUsers[u.(string)] = true
		}
	}
	for u := range usersRemoved {

		_, isOrgUser := orgUsers[u]
		_, isSpaceUser := usersAdded[u]

		if !isOrgUser && !isSpaceUser {

			session.Log.DebugMessage(
				"Removing user '%s' from org '%s' as he/she no longer has an assigned role within the org.",
				u, orgID)

			if err = om.RemoveUser(orgID, u, cfapi.OrgRoleMember); err != nil {
				return err
			}
		}
	}

	segID := d.Get("isolation_segment_id").(string)
	err = sm.SetSpaceSegment(spaceID, segID)
	if err != nil {
		return err
	}

	return nil
}

func resourceSpaceDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	err = session.SpaceManager().DeleteSpace(d.Id())
	return err
}
