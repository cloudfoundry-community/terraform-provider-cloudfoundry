package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOrg() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceOrgCreate,
		ReadContext:   resourceOrgRead,
		UpdateContext: resourceOrgUpdate,
		DeleteContext: resourceOrgDelete,

		Importer: &schema.ResourceImporter{
			StateContext: ImportReadContext(resourceOrgRead),
		},

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
				Deprecated: "Use resource cloudfoundry_org_users instead",
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			"billing_managers": &schema.Schema{
				Deprecated: "Use resource cloudfoundry_org_users instead",
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			"auditors": &schema.Schema{
				Deprecated: "Use resource cloudfoundry_org_users instead",
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
			"delete_recursive_allowed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Allow recursive deletion of spaces.",
			},
		},
	}
}

func resourceOrgCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	om := session.ClientV2

	name := d.Get("name").(string)
	quota := d.Get("quota").(string)

	org, _, err := om.CreateOrganization(name, quota)
	if err != nil {
		return diag.FromErr(err)
	}
	if quota == "" {
		d.Set("quota", org.QuotaDefinitionGUID)
	}
	d.SetId(org.GUID)
	return resourceOrgUpdate(ctx, d, meta)
}

func resourceOrgRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	om := session.ClientV2

	id := d.Id()

	org, _, err := om.GetOrganization(id)
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("name", org.Name)
	d.Set("quota", org.QuotaDefinitionGUID)

	for t, r := range orgRoleMap {
		users, _, err := om.GetOrganizationUsersByRole(r, id)
		if err != nil {
			return diag.FromErr(err)
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
	err = metadataRead(orgMetadata, d, meta, false)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceOrgUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	id := d.Id()
	om := session.ClientV2

	if !d.IsNewResource() {
		_, _, err := om.UpdateOrganization(id, d.Get("name").(string), d.Get("quota").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	for t, r := range orgRoleMap {
		remove, add := getListChanges(d.GetChange(t))

		for _, uid := range remove {
			_, err := om.DeleteOrganizationUserByRole(r, id, uid)
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
			err = updateOrgUserByRole(session, r, id, uidOrUsername, byUsername)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	err := metadataUpdate(orgMetadata, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceOrgDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	client := session.ClientV3

	id := d.Id()
	spaces, _, _, err := session.ClientV3.GetSpaces(ccv3.Query{
		Key:    ccv3.OrganizationGUIDFilter,
		Values: []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !session.Config.DeleteRecursiveAllowed && len(spaces) > 0 {
		return diag.Errorf("Organization %s has %d spaces. Please delete them first or set delete_recursive_allowed to true", id, len(spaces))
	}
	for _, s := range spaces {
		j, _, err := client.DeleteSpace(s.GUID)
		if err != nil {
			return diag.FromErr(err)
		}
		_, err = client.PollJob(j)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	j, _, err := client.DeleteOrganization(id)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.PollJob(j)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
