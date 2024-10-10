package cloudfoundry

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

var orgRoleMap = map[string]constant.UserRole{
	"managers":         constant.OrgManager,
	"billing_managers": constant.BillingManager,
	"auditors":         constant.OrgAuditor,
}

func resourceOrgUsers() *schema.Resource {

	return &schema.Resource{

		CreateContext: resourceOrgUsersCreate,
		ReadContext:   resourceOrgUsersRead,
		UpdateContext: resourceOrgUsersUpdate,
		DeleteContext: resourceOrgUsersDelete,

		Importer: &schema.ResourceImporter{
			StateContext: ImportReadContext(resourceOrgUsersRead),
		},

		Schema: map[string]*schema.Schema{
			"org": &schema.Schema{
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
			"billing_managers": &schema.Schema{
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

func resourceOrgUsersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	session := meta.(*managers.Session)
	orgId := d.Get("org").(string)
	d.SetId(id)
	if d.Get("force").(bool) {
		for _, r := range orgRoleMap {
			users, _, err := session.ClientV2.GetOrganizationUsersByRole(r, orgId)
			if err != nil {
				return diag.FromErr(err)
			}
			for _, u := range users {
				_, err := session.ClientV2.DeleteOrganizationUserByRole(r, orgId, u.GUID)
				if err != nil {
					return diag.FromErr(err)
				}
			}

		}
	}
	return resourceOrgUsersUpdate(ctx, d, meta)
}

func resourceOrgUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if IsImportState(d) {
		d.Set("org", d.Id())
	}
	session := meta.(*managers.Session)
	for t, r := range orgRoleMap {
		users, _, err := session.ClientV2.GetOrganizationUsersByRole(r, d.Get("org").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		tfUsers := d.Get(t).(*schema.Set).List()
		if d.Get("force").(bool) || IsImportState(d) {
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
		} else {
			finalUsers := intersectSlices(tfUsers, users, func(source, item interface{}) bool {
				return source.(string) == item.(ccv2.User).GUID || strings.EqualFold(source.(string), item.(ccv2.User).Username)
			})
			d.Set(t, schema.NewSet(resourceStringHash, finalUsers))
		}
	}
	return nil
}

func resourceOrgUsersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	orgId := d.Get("org").(string)
	for t, r := range orgRoleMap {
		remove, add := getListChanges(d.GetChange(t))
		for _, uid := range remove {
			byUsername := true
			_, err := uuid.ParseUUID(uid)
			if err == nil {
				byUsername = false
			}
			err = deleteOrgUserByRole(session, r, orgId, uid, byUsername)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		for _, uid := range add {
			byUsername := true
			_, err := uuid.ParseUUID(uid)
			if err == nil {
				byUsername = false
			}
			err = updateOrgUserByRole(session, r, orgId, uid, byUsername)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func resourceOrgUsersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	orgId := d.Get("org").(string)
	session := meta.(*managers.Session)
	for t, r := range orgRoleMap {
		tfUsers := d.Get(t).(*schema.Set).List()
		for _, uid := range tfUsers {
			byUsername := true
			_, err := uuid.ParseUUID(uid.(string))
			if err == nil {
				byUsername = false
			}

			err = deleteOrgUserByRole(session, r, orgId, uid.(string), byUsername)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func updateOrgUserByRole(session *managers.Session, role constant.UserRole, guid string, guidOrUsername string, byUsername bool) error {
	if !byUsername {
		_, err := session.ClientV2.UpdateOrganizationUserByRole(role, guid, guidOrUsername)
		if err != nil {
			return err
		}
		return nil
	}
	var err error
	switch role {
	case constant.OrgAuditor:
		_, err = session.ClientV2.UpdateOrganizationAuditorByUsername(guid, guidOrUsername)
	case constant.OrgManager:
		_, err = session.ClientV2.UpdateOrganizationManagerByUsername(guid, guidOrUsername)
	case constant.BillingManager:
		_, err = session.ClientV2.UpdateOrganizationBillingManagerByUsername(guid, guidOrUsername)
	default:
		_, err = session.ClientV2.UpdateOrganizationUserByUsername(guid, guidOrUsername)
	}
	return err
}

func deleteOrgUserByRole(session *managers.Session, role constant.UserRole, guid string, guidOrUsername string, byUsername bool) error {
	if !byUsername {
		_, err := session.ClientV2.DeleteOrganizationUserByRole(role, guid, guidOrUsername)
		if err != nil {
			return err
		}
		return nil
	}
	var err error
	switch role {
	case constant.OrgAuditor:
		_, err = session.ClientV2.DeleteOrganizationAuditorByUsername(guid, guidOrUsername)
	case constant.OrgManager:
		_, err = session.ClientV2.DeleteOrganizationManagerByUsername(guid, guidOrUsername)
	case constant.BillingManager:
		_, err = session.ClientV2.DeleteOrganizationBillingManagerByUsername(guid, guidOrUsername)
	default:
		_, err = session.ClientV2.DeleteOrganizationUserByUsername(guid, guidOrUsername)
	}
	return err
}
