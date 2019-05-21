package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

var typeToSpaceRoleMap = map[string]constant.UserRole{
	"managers":   constant.SpaceManager,
	"developers": constant.SpaceDeveloper,
	"auditors":   constant.SpaceAuditor,
}

func resourceSpaceUsers() *schema.Resource {

	return &schema.Resource{

		Create: resourceSpaceUsersCreate,
		Read:   resourceSpaceUsersRead,
		Update: resourceSpaceUsersUpdate,
		Delete: resourceSpaceUsersDelete,

		Schema: map[string]*schema.Schema{
			"space": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"force": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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

func resourceSpaceUsersCreate(d *schema.ResourceData, meta interface{}) error {
	id, err := uuid.GenerateUUID()
	if err != nil {
		return err
	}
	session := meta.(*managers.Session)
	spaceId := d.Get("space").(string)
	d.SetId(id)
	if d.Get("force").(bool) {
		for _, r := range typeToSpaceRoleMap {
			users, _, err := session.ClientV2.GetSpaceUsersByRole(r, spaceId)
			if err != nil {
				return err
			}
			for _, u := range users {
				_, err := session.ClientV2.DeleteSpaceUserByRole(r, spaceId, u.GUID)
				if err != nil {
					return err
				}
			}

		}
	}
	return resourceSpaceUsersUpdate(d, meta)
}

func resourceSpaceUsersRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	for t, r := range typeToSpaceRoleMap {
		users, _, err := session.ClientV2.GetSpaceUsersByRole(r, d.Get("space").(string))
		if err != nil {
			return err
		}
		tfUsers := d.Get(t).(*schema.Set).List()
		if !d.Get("force").(bool) {
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
	return nil
}

func resourceSpaceUsersUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	spaceId := d.Get("space").(string)
	space, _, err := session.ClientV2.GetSpace(spaceId)
	if err != nil {
		return err
	}
	for t, r := range typeToSpaceRoleMap {
		remove, add := getListChanges(d.GetChange(t))
		for _, uid := range remove {
			_, err = session.ClientV2.DeleteSpaceUserByRole(r, spaceId, uid)
			if err != nil {
				return err
			}
		}
		for _, uid := range add {
			err = addOrNothingUserInOrgBySpace(session.ClientV2, space.OrganizationGUID, uid)
			if err != nil {
				return err
			}
			_, err = session.ClientV2.UpdateSpaceUserByRole(r, spaceId, uid)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func addOrNothingUserInOrgBySpace(client *ccv2.Client, orgId, uaaid string) error {
	orgs, _, err := client.GetUserOrganizations(uaaid)
	_, isNotFound := err.(ccerror.ResourceNotFoundError)
	if err != nil && !isNotFound {
		return err
	}

	if isInSlice(orgs, func(object interface{}) bool {
		return object.(ccv2.Organization).GUID == orgId
	}) && !isNotFound {
		return nil
	}
	_, err = client.UpdateOrganizationUserByRole(constant.OrgUser, orgId, uaaid)
	return err
}

func resourceSpaceUsersDelete(d *schema.ResourceData, meta interface{}) error {
	spaceId := d.Get("space").(string)
	session := meta.(*managers.Session)
	for t, r := range typeToSpaceRoleMap {
		tfUsers := d.Get(t).(*schema.Set).List()
		for _, uid := range tfUsers {
			_, err := session.ClientV2.DeleteSpaceUserByRole(r, spaceId, uid.(string))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
