package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/uaa"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: CaseDifference,
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: CaseDifference,
			},
			"family_name": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: CaseDifference,
			},
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"groups": &schema.Schema{
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Optional:   true,
				Computed:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
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
	uaam := session.ClientUAA
	um := session.ClientV2

	userUAA, err := createUaaUserIfNotExists(username, password, origin, &name, emails, uaam)
	if err != nil {
		return err
	}

	userCF, err := createCFUserIfNotExists(userUAA.ID, um)
	if err != nil {
		return err
	}

	d.SetId(userCF.GUID)
	return resourceUserUpdate(d, meta)
}

func createCFUserIfNotExists(ID string, um *ccv2.Client) (ccv2.User, error) {
	users, _, err := um.GetUsers()
	if err != nil {
		return ccv2.User{}, err
	}

	for _, user := range users {
		if user.GUID == ID {
			return user, nil
		}
	}

	user, _, err := um.CreateUser(ID)
	return user, err
}

func createUaaUserIfNotExists(
	username string,
	password string,
	origin string,
	name *uaa.UserName,
	emails []uaa.Email,
	uaam *uaa.Client,
) (uaa.User, error) {

	createUser := func() (uaa.User, error) {
		return uaam.CreateUserFromObject(uaa.User{
			Username: username,
			Password: password,
			Origin:   origin,
			Name:     *name,
			Emails:   emails,
		})
	}

	// create uaa user if not exists
	users, err := uaam.GetUsersByUsername(username)
	if err != nil {
		if IsErrNotFound(err) {
			return createUser()
		}
		return uaa.User{}, err
	}

	for _, uaaUser := range users {
		// warn: GetUsersByUsername only fetch id and username attributes
		user, err := uaam.GetUser(uaaUser.ID)
		if err != nil {
			return uaa.User{}, err
		}
		if user.Origin == origin {
			return user, nil
		}
	}
	return createUser()
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)

	umuaa := session.ClientUAA
	um := session.ClientV2
	id := d.Id()

	// 1. check  user exists in CF
	usersCF, _, err := um.GetUsers()
	if err != nil {
		d.SetId("")
		return err
	}
	found := false
	for _, user := range usersCF {
		if user.GUID == id {
			found = true
			break
		}
	}
	if !found {
		d.SetId("")
		return nil
	}

	// 2. check  user exists in UAA
	user, err := umuaa.GetUser(id)
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", user.Username)
	d.Set("origin", user.Origin)
	d.Set("given_name", user.Name.GivenName)
	d.Set("family_name", user.Name.FamilyName)
	d.Set("email", user.Emails[0].Value)

	tfGroups := d.Get("groups").(*schema.Set).List()
	groups := user.Groups
	if !IsImportState(d) {
		finalGroups := intersectSlices(tfGroups, groups, func(source, item interface{}) bool {
			return source.(string) == item.(uaa.Group).Name()
		})
		d.Set("groups", schema.NewSet(resourceStringHash, finalGroups))
	} else {
		d.Set("groups", schema.NewSet(resourceStringHash, objectsToIds(groups, func(object interface{}) string {
			return object.(uaa.Group).Name()
		})))
	}
	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	id := d.Id()
	umuaa := session.ClientUAA

	if !d.IsNewResource() {

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
			_, err := umuaa.UpdateUser(uaa.User{
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
			err := umuaa.ChangeUserPassword(id, oldPassword, newPassword)
			if err != nil {
				return err
			}
		}
	}

	old, new := d.GetChange("groups")
	rolesToDelete, rolesToAdd := getListChanges(old, new)

	for _, r := range rolesToDelete {
		err := umuaa.DeleteMemberByName(id, r)
		if err != nil {
			return err
		}
	}

	for _, r := range rolesToAdd {
		err := umuaa.AddMemberByName(id, d.Get("origin").(string), r)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	id := d.Id()
	umuaa := session.ClientUAA
	return umuaa.DeleteUser(id)
}
