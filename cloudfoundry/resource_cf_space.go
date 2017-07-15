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
			"asgs": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"managers": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"developers": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"auditors": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
		},
	}
}

var typeToSpaceRoleMap = map[string]cfapi.SpaceRole{
	"managers":   cfapi.SpaceRoleManager,
	"developers": cfapi.SpaceRoleDeveloper,
	"auditors":   cfapi.SpaceRoleAuditor,
}

func resourceSpaceCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		name, org, quota string
		allowSSH         bool
		asgs             []interface{}
	)
	name = d.Get("name").(string)
	org = d.Get("org").(string)
	if v, ok := d.GetOk("quota"); ok {
		quota = v.(string)
	}
	if v, ok := d.GetOk("asgs"); ok {
		asgs = v.(*schema.Set).List()
	}
	allowSSH = d.Get("allow_ssh").(bool)

	var id string

	sm := session.SpaceManager()
	if id, err = sm.CreateSpace(name, org, quota, allowSSH, asgs); err != nil {
		return err
	}
	d.SetId(id)
	return resourceSpaceUpdate(d, NewResourceMeta{meta})
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

		runningAsgs     []string
		spaceAsgs, asgs []interface{}
	)

	if space, err = sm.ReadSpace(id); err != nil {
		return
	}
	d.Set("name", space.Name)
	d.Set("org", space.OrgGUID)
	d.Set("quota", space.QuotaGUID)
	d.Set("allow_ssh", space.AllowSSH)

	var users []interface{}
	for t, r := range typeToSpaceRoleMap {
		if users, err = sm.ListUsers(id, r); err != nil {
			return
		}
		d.Set(t, schema.NewSet(resourceStringHash, users))
	}

	if runningAsgs, err = session.ASGManager().Running(); err != nil {
		return err
	}
	if spaceAsgs, err = sm.ListASGs(id); err != nil {
		return
	}
	for _, a := range spaceAsgs {
		if !isStringInList(runningAsgs, a.(string)) {
			asgs = append(asgs, a)
		}
	}
	d.Set("asgs", schema.NewSet(resourceStringHash, asgs))

	return
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
	orgID := d.Get("org").(string)

	om := session.OrgManager()
	sm := session.SpaceManager()

	if !newResource {

		var asgs []interface{}

		space := cfapi.CCSpace{
			ID:   spaceID,
			Name: d.Get("name").(string),

			OrgGUID: orgID,

			AllowSSH: d.Get("allow_ssh").(bool),
		}
		if v, ok := d.GetOk("quota"); ok {
			space.QuotaGUID = v.(string)
		}
		if v, ok := d.GetOk("asgs"); ok {
			asgs = v.(*schema.Set).List()
		}

		if err = sm.UpdateSpace(space, asgs); err != nil {
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
				return
			}
		}
		for _, uid := range add {
			session.Log.DebugMessage("Adding user '%s' to space '%s' with role '%s'.", uid, spaceID, r)
			if err = om.AddUser(orgID, uid, cfapi.OrgRoleMember); err != nil {
				return
			}
			if err = sm.AddUser(spaceID, uid, r); err != nil {
				return
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
			return
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
				return
			}
		}
	}

	return
}

func resourceSpaceDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	err = session.SpaceManager().DeleteSpace(d.Id())
	return
}
