package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

func resourceOrg() *schema.Resource {

	return &schema.Resource{

		Create: resourceOrgCreate,
		Read:   resourceOrgRead,
		Update: resourceOrgUpdate,
		Delete: resourceOrgDelete,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"quota": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"managers": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"billing_managers": &schema.Schema{
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

var orgRoleMap = map[string]cfapi.OrgRole{
	"managers":         cfapi.OrgRoleManager,
	"billing_managers": cfapi.OrgRoleBillingManager,
	"auditors":         cfapi.OrgRoleAuditor,
}

func resourceOrgCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		name, quota string
		org         cfapi.CCOrg
	)
	name = d.Get("name").(string)
	if v, ok := d.GetOk("quota"); ok {
		quota = v.(string)
	}

	om := session.OrgManager()
	if org, err = om.CreateOrg(name, quota); err != nil {
		return err
	}
	if len(quota) == 0 {
		d.Set("quota", org.QuotaGUID)
	}
	d.SetId(org.ID)
	return resourceOrgUpdate(d, NewResourceMeta{meta})
}

func resourceOrgRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id := d.Id()
	om := session.OrgManager()

	var org cfapi.CCOrg
	if org, err = om.ReadOrg(id); err != nil {
		return
	}

	d.Set("name", org.Name)
	d.Set("quota", org.QuotaGUID)

	var users []interface{}
	for t, r := range orgRoleMap {
		if users, err = om.ListUsers(id, r); err != nil {
			return
		}
		d.Set(t, schema.NewSet(resourceStringHash, users))
	}

	return
}

func resourceOrgUpdate(d *schema.ResourceData, meta interface{}) (err error) {

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

	id := d.Id()
	om := session.OrgManager()

	if !newResource {

		org := cfapi.CCOrg{
			ID:   id,
			Name: d.Get("name").(string),
		}
		if v, ok := d.GetOk("quota"); ok {
			org.QuotaGUID = v.(string)
		}

		if err = om.UpdateOrg(org); err != nil {
			return err
		}
	}

	for t, r := range orgRoleMap {
		old, new := d.GetChange(t)
		remove, add := getListChanges(old, new)

		for _, uid := range remove {
			session.Log.DebugMessage("Removing user '%s' from organization '%s' with role '%s'.", uid, id, r)
			if err = om.RemoveUser(id, uid, r); err != nil {
				return
			}
		}
		for _, uid := range add {
			session.Log.DebugMessage("Adding user '%s' to organization '%s' with role '%s'.", uid, id, r)
			if err = om.AddUser(id, uid, r); err != nil {
				return
			}
		}
	}
	return
}

func resourceOrgDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	om := session.OrgManager()
	sm := session.SpaceManager()

	id := d.Id()

	var spaces []cfapi.CCSpace
	if spaces, err = sm.FindSpacesInOrg(id); err != nil {
		return
	}
	for _, s := range spaces {
		if err = sm.DeleteSpace(s.ID); err != nil {
			return
		}
	}

	err = om.DeleteOrg(d.Id())
	return
}
