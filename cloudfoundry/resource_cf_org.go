package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"github.com/hashicorp/go-uuid"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOrg() *schema.Resource {
	return &schema.Resource{

		Create: resourceOrgCreate,
		Read:   resourceOrgRead,
		Update: resourceOrgUpdate,
		Delete: resourceOrgDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceOrgRead),
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
		},
	}
}

func resourceOrgCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	om := session.ClientV2

	name := d.Get("name").(string)
	quota := d.Get("quota").(string)

	org, _, err := om.CreateOrganization(name, quota)
	if err != nil {
		return err
	}
	if quota == "" {
		d.Set("quota", org.QuotaDefinitionGUID)
	}
	d.SetId(org.GUID)
	return resourceOrgUpdate(d, meta)
}

func resourceOrgRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	om := session.ClientV2

	id := d.Id()

	org, _, err := om.GetOrganization(id)
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", org.Name)
	d.Set("quota", org.QuotaDefinitionGUID)

	for t, r := range orgRoleMap {
		users, _, err := om.GetOrganizationUsersByRole(r, id)
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
	err = metadataRead(orgMetadata, d, meta, false)
	if err != nil {
		return err
	}
	return nil
}

func resourceOrgUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*managers.Session)

	id := d.Id()
	om := session.ClientV2

	if !d.IsNewResource() {
		_, _, err := om.UpdateOrganization(id, d.Get("name").(string), d.Get("quota").(string))
		if err != nil {
			return err
		}
	}

	for t, r := range orgRoleMap {
		remove, add := getListChanges(d.GetChange(t))

		for _, uid := range remove {
			_, err = om.DeleteOrganizationUserByRole(r, id, uid)
			if err != nil {
				return err
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
				return err
			}
		}
	}
	err = metadataUpdate(orgMetadata, d, meta)
	if err != nil {
		return err
	}
	return nil
}

func resourceOrgDelete(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*managers.Session)
	client := session.ClientV2

	id := d.Id()
	spaces, _, err := client.GetSpaces(ccv2.FilterByOrg(id))

	if err != nil {
		return err
	}
	for _, s := range spaces {
		j, _, err := client.DeleteSpace(s.GUID)
		if err != nil {
			return err
		}
		_, err = client.PollJob(j)
		if err != nil {
			return err
		}
	}
	j, _, err := client.DeleteOrganization(id)
	if err != nil {
		return err
	}
	_, err = client.PollJob(j)
	if err != nil {
		return err
	}
	return nil
}
