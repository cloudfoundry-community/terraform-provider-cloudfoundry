package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/uaa"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{

		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceUserRead),
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"origin": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "uaa",
			},
			"given_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"family_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"groups": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
		},
	}
}

func resourceUserCreate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	username := d.Get("name").(string)
	password := d.Get("password").(string)
	origin := d.Get("origin").(string)
	givenName := d.Get("given_name").(string)
	familyName := d.Get("family_name").(string)

	email := d.Get("email").(string)
	if email == "" {
		email = username
		d.Set("email", email)
	}
	emails := make([]uaa.Email, 0)
	if email != "" {
		emails = append(emails, uaa.Email{
			Value:   email,
			Primary: true,
		})
	}

	var name uaa.UserName
	if givenName != "" || familyName != "" {
		name = uaa.UserName{
			GivenName:  givenName,
			FamilyName: familyName,
		}
	}
	um := session.ClientUAA
	user, err := um.CreateUserFromObject(uaa.User{
		Username: username,
		Password: password,
		Origin:   origin,
		Name:     name,
		Emails:   emails,
	})
	if err != nil {
		return err
	}

	d.SetId(user.ID)
	return resourceUserUpdate(d, NewResourceMeta{meta})
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)

	um := session.ClientUAA
	id := d.Id()

	user, err := um.GetUser(id)
	if err != nil {
		d.SetId("")
		return err
	}

	d.Set("name", user.Username)
	d.Set("origin", user.Origin)
	d.Set("given_name", user.Name.GivenName)
	d.Set("family_name", user.Name.FamilyName)
	d.Set("email", user.Emails[0].Value)

	var groups []interface{}
	for _, g := range user.Groups {
		if !session.IsUaaDefaultCfGroup(g.Name()) {
			groups = append(groups, g.Name())
		}
	}
	d.Set("groups", schema.NewSet(resourceStringHash, groups))

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {

	var (
		newResource bool
		session     *managers.Session
	)

	if m, ok := meta.(NewResourceMeta); ok {
		session = m.meta.(*managers.Session)
		newResource = true
	} else {
		session = meta.(*managers.Session)
		if session == nil {
			return fmt.Errorf("client is nil")
		}
		newResource = false
	}

	id := d.Id()
	um := session.ClientUAA

	if !newResource {

		updateUserDetail := false
		name := d.Get("name").(string)
		if d.HasChange("name") {
			updateUserDetail = true
		}
		givenName := d.Get("given_name").(string)
		if d.HasChange("given_name") {
			updateUserDetail = true
		}
		familyName := d.Get("family_name").(string)
		if d.HasChange("family_name") {
			updateUserDetail = true
		}

		var username uaa.UserName
		if givenName != "" || familyName != "" {
			username = uaa.UserName{
				GivenName:  givenName,
				FamilyName: familyName,
			}
		}

		emails := []uaa.Email{
			{
				Value:   d.Get("email").(string),
				Primary: true,
			},
		}

		if d.HasChange("email") {
			updateUserDetail = true
		}

		if updateUserDetail {
			_, err := um.UpdateUser(uaa.User{
				ID:       id,
				Username: name,
				Emails:   emails,
				Name:     username,
				Origin:   d.Get("origin").(string),
			})
			if err != nil {
				return err
			}
		}

		updatePassword, oldPassword, newPassword := getResourceChange("password", d)
		if updatePassword {
			err := um.ChangeUserPassword(id, oldPassword, newPassword)
			if err != nil {
				return err
			}
		}
	}

	old, new := d.GetChange("groups")
	rolesToDelete, rolesToAdd := getListChanges(old, new)

	for _, r := range rolesToDelete {
		err := um.DeleteMemberByName(id, r)
		if err != nil {
			return err
		}
	}

	for _, r := range rolesToAdd {
		err := um.AddMemberByName(id, d.Get("origin").(string), r)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	id := d.Id()
	um := session.ClientUAA
	return um.DeleteUser(id)
}
