package cloudfoundry

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

var typeToSpaceRoleMap = map[string]constant.UserRole{
	"managers":   constant.SpaceManager,
	"developers": constant.SpaceDeveloper,
	"auditors":   constant.SpaceAuditor,
}

func resourceSpaceUsers() *schema.Resource {

	return &schema.Resource{

		CreateContext: resourceSpaceUsersCreate,
		ReadContext:   resourceSpaceUsersRead,
		UpdateContext: resourceSpaceUsersUpdate,
		DeleteContext: resourceSpaceUsersDelete,

		Importer: &schema.ResourceImporter{
			StateContext: ImportReadContext(resourceSpaceUsersRead),
		},

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
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			"developers": &schema.Schema{
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			"auditors": &schema.Schema{
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
		},
	}
}

func resourceSpaceUsersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	session := meta.(*managers.Session)
	spaceId := d.Get("space").(string)
	d.SetId(id)
	if d.Get("force").(bool) {
		for _, r := range typeToSpaceRoleMap {
			users, _, err := session.ClientV2.GetSpaceUsersByRole(r, spaceId)
			if err != nil {
				return diag.FromErr(err)
			}
			for _, u := range users {
				_, err := session.ClientV2.DeleteSpaceUserByRole(r, spaceId, u.GUID)
				if err != nil {
					return diag.FromErr(err)
				}
			}

		}
	}
	return resourceSpaceUsersUpdate(ctx, d, meta)
}

func resourceSpaceUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if IsImportState(d) {
		d.Set("space", d.Id())
		d.Set("force", false)
	}
	session := meta.(*managers.Session)
	for t, r := range typeToSpaceRoleMap {
		users, _, err := session.ClientV2.GetSpaceUsersByRole(r, d.Get("space").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		tfUsers := d.Get(t).(*schema.Set).List()
		if !d.Get("force").(bool) && !IsImportState(d) {
			finalUsers := intersectSlices(tfUsers, users, func(source, item interface{}) bool {
				return source.(string) == item.(ccv2.User).GUID || strings.EqualFold(source.(string), item.(ccv2.User).Username)
			})
			d.Set(t, schema.NewSet(resourceStringHash, finalUsers))
		} else {
			usersByUsername := intersectSlices(tfUsers, users, func(source, item interface{}) bool {
				return strings.EqualFold(source.(string), item.(ccv2.User).Username)
			})

			d.Set(t, schema.NewSet(resourceStringHash, objectsToIds(users, func(object interface{}) string {
				if isInSlice(usersByUsername, func(userByUsername interface{}) bool {
					return strings.EqualFold(object.(ccv2.User).Username, userByUsername.(string))
				}) {
					return object.(ccv2.User).Username
				}
				return object.(ccv2.User).GUID
			})))
		}
	}
	return nil
}

func resourceSpaceUsersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	spaceId := d.Get("space").(string)
	space, _, err := session.ClientV2.GetSpace(spaceId)
	if err != nil {
		return diag.FromErr(err)
	}
	for t, r := range typeToSpaceRoleMap {
		remove, add := getListChanges(d.GetChange(t))
		for _, uid := range remove {
			byUsername := true
			_, err = uuid.ParseUUID(uid)
			if err == nil {
				byUsername = false
			}

			err = deleteSpaceUserByRole(session, r, spaceId, uid, byUsername)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		for _, uidOrUsername := range add {
			byUsername := true
			_, err := uuid.ParseUUID(uidOrUsername)
			if err == nil {
				byUsername = false
			}
			err = addOrNothingUserInOrgBySpace(session, space.OrganizationGUID, uidOrUsername, byUsername)
			if err != nil {
				return diag.FromErr(err)
			}

			err = updateSpaceUserByRole(session, r, spaceId, uidOrUsername, byUsername)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func updateSpaceUserByRole(session *managers.Session, role constant.UserRole, guid string, guidOrUsername string, byUsername bool) error {
	if !byUsername {
		_, err := session.ClientV2.UpdateSpaceUserByRole(role, guid, guidOrUsername)
		if err != nil {
			return err
		}
		return nil
	}
	var err error
	switch role {
	case constant.SpaceAuditor:
		_, err = session.ClientV2.UpdateSpaceAuditorByUsername(guid, guidOrUsername)
	case constant.SpaceManager:
		_, err = session.ClientV2.UpdateSpaceManagerByUsername(guid, guidOrUsername)
	default:
		_, err = session.ClientV2.UpdateSpaceDeveloperByUsername(guid, guidOrUsername)
	}
	return err
}

func deleteSpaceUserByRole(session *managers.Session, role constant.UserRole, guid string, guidOrUsername string, byUsername bool) error {
	if !byUsername {
		_, err := session.ClientV2.DeleteSpaceUserByRole(role, guid, guidOrUsername)
		if err != nil {
			return err
		}
		return nil
	}
	var err error
	switch role {
	case constant.SpaceAuditor:
		_, err = session.ClientV2.DeleteSpaceAuditorByUsername(guid, guidOrUsername)
	case constant.SpaceManager:
		_, err = session.ClientV2.DeleteSpaceManagerByUsername(guid, guidOrUsername)
	default:
		_, err = session.ClientV2.DeleteSpaceDeveloperByUsername(guid, guidOrUsername)
	}
	return err
}

func addOrNothingUserInOrgBySpace(session *managers.Session, orgId, uaaidOrUsername string, byUsername bool) error {
	client := session.ClientV2
	orgs, _, err := client.GetUserOrganizations(uaaidOrUsername)
	isNotFound := IsErrNotFound(err)
	isNotAuthorized := IsErrNotAuthorized(err)
	if err != nil && !isNotFound && !isNotAuthorized {
		return err
	}

	if !isNotAuthorized && isInSlice(orgs, func(object interface{}) bool {
		return object.(ccv2.Organization).GUID == orgId
	}) && !isNotFound {
		return nil
	}

	// Fallback for isNotAuthorized case
	if isNotAuthorized {
		users, _, err := client.GetOrganizationUsers(orgId)
		if err != nil {
			return err
		}
		if isInSlice(users, func(object interface{}) bool {
			return object.(ccv2.User).GUID == uaaidOrUsername
		}) {
			return nil
		}
	}
	return updateOrgUserByRole(session, constant.OrgUser, orgId, uaaidOrUsername, byUsername)
}

func resourceSpaceUsersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	spaceId := d.Get("space").(string)
	session := meta.(*managers.Session)
	for t, r := range typeToSpaceRoleMap {
		tfUsers := d.Get(t).(*schema.Set).List()
		for _, uid := range tfUsers {
			uaaIDOrUsername := uid.(string)
			byUsername := true
			_, err := uuid.ParseUUID(uaaIDOrUsername)
			if err == nil {
				byUsername = false
			}

			err = deleteSpaceUserByRole(session, r, spaceId, uaaIDOrUsername, byUsername)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}
