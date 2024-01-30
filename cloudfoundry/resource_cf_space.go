package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceSpace() *schema.Resource {

	return &schema.Resource{

		CreateContext: resourceSpaceCreate,
		ReadContext:   resourceSpaceRead,
		UpdateContext: resourceSpaceUpdate,
		DeleteContext: resourceSpaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: ImportReadContext(resourceSpaceRead),
		},

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
				Computed: true,
			},
			"isolation_segment": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"asgs": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"staging_asgs": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
			"managers": &schema.Schema{
				Deprecated: "Use resource cloudfoundry_space_users instead",
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			"developers": &schema.Schema{
				Deprecated: "Use resource cloudfoundry_space_users instead",
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			"auditors": &schema.Schema{
				Deprecated: "Use resource cloudfoundry_space_users instead",
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
				Description: "Allow recursive deletion of apps, routes, service instances.",
			},
		},
	}
}

func resourceSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	org := d.Get("org").(string)
	quota := d.Get("quota").(string)
	allowSSH := true
	// if user does explicitly set allow_ssh
	// it set allow the user value
	if allow, ok := d.GetOk("allow_ssh"); ok {
		allowSSH = allow.(bool)
	}

	sm := session.ClientV2
	space, _, err := sm.CreateSpaceFromObject(ccv2.Space{
		Name:                     name,
		OrganizationGUID:         org,
		AllowSSH:                 allowSSH,
		SpaceQuotaDefinitionGUID: quota,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(space.GUID)
	dg := resourceSpaceUpdate(ctx, d, meta)
	if dg.HasError() {
		return dg
	}
	err = metadataCreate(spaceMetadata, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	id := d.Id()
	sm := session.ClientV2

	space, _, err := sm.GetSpace(id)
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	d.Set("name", space.Name)
	d.Set("org", space.OrganizationGUID)
	d.Set("quota", space.SpaceQuotaDefinitionGUID)
	d.Set("allow_ssh", space.AllowSSH)

	for t, r := range typeToSpaceRoleMap {
		users, _, err := sm.GetSpaceUsersByRole(r, id)
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

	runningAsgs, _, err := session.ClientV2.GetSpaceSecurityGroups(d.Id())
	if err != nil {
		return nil
	}
	if !IsImportState(d) {
		finalRunningAsg := intersectSlices(d.Get("asgs").(*schema.Set).List(), runningAsgs, func(source, item interface{}) bool {
			return source.(string) == item.(ccv2.SecurityGroup).GUID
		})
		d.Set("asgs", schema.NewSet(resourceStringHash, finalRunningAsg))
	} else {
		finalRunningAsgs, _ := getInSlice(runningAsgs, func(object interface{}) bool {
			return !object.(ccv2.SecurityGroup).RunningDefault
		})
		d.Set("asgs", schema.NewSet(resourceStringHash, objectsToIds(finalRunningAsgs, func(object interface{}) string {
			return object.(ccv2.SecurityGroup).GUID
		})))
	}

	stagingAsgs, _, err := session.ClientV2.GetSpaceStagingSecurityGroups(d.Id())
	if err != nil {
		return nil
	}
	if !IsImportState(d) {
		finalStagingAsg := intersectSlices(d.Get("staging_asgs").(*schema.Set).List(), stagingAsgs, func(source, item interface{}) bool {
			return source.(string) == item.(ccv2.SecurityGroup).GUID
		})
		d.Set("staging_asgs", schema.NewSet(resourceStringHash, finalStagingAsg))
	} else {
		finalStagingAsgs, _ := getInSlice(stagingAsgs, func(object interface{}) bool {
			return !object.(ccv2.SecurityGroup).StagingDefault
		})
		d.Set("staging_asgs", schema.NewSet(resourceStringHash, objectsToIds(finalStagingAsgs, func(object interface{}) string {
			return object.(ccv2.SecurityGroup).GUID
		})))
	}

	segment, _, err := session.ClientV3.GetSpaceIsolationSegment(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("isolation_segment", segment.GUID)

	err = metadataRead(spaceMetadata, d, meta, false)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	spaceID := d.Id()
	orgID := d.Get("org").(string)
	if !d.IsNewResource() {
		_, _, err := session.ClientV2.UpdateSpace(ccv2.Space{
			GUID:             spaceID,
			Name:             d.Get("name").(string),
			OrganizationGUID: orgID,
			AllowSSH:         d.Get("allow_ssh").(bool),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if d.HasChange("quota") {
			_, err := session.ClientV2.SetSpaceQuota(spaceID, d.Get("quota").(string))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	var err error
	removeAsgs, addAsgs := getListChanges(d.GetChange("asgs"))
	for _, asgID := range removeAsgs {
		_, err = session.ClientV2.DeleteSecurityGroupSpace(asgID, spaceID)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	for _, asgID := range addAsgs {
		_, err = session.ClientV2.UpdateSecurityGroupSpace(asgID, spaceID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	removeStagingAsgs, addStagingAsgs := getListChanges(d.GetChange("staging_asgs"))
	for _, asgID := range removeStagingAsgs {
		_, err = session.ClientV2.DeleteSecurityGroupStagingSpace(asgID, spaceID)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	for _, asgID := range addStagingAsgs {
		_, err = session.ClientV2.UpdateSecurityGroupStagingSpace(asgID, spaceID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	for t, r := range typeToSpaceRoleMap {
		remove, add := getListChanges(d.GetChange(t))
		for _, uid := range remove {
			_, err = session.ClientV2.DeleteSpaceUserByRole(r, spaceID, uid)
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
			err = addOrNothingUserInOrgBySpace(session, orgID, uidOrUsername, byUsername)
			if err != nil {
				return diag.FromErr(err)
			}
			err = updateSpaceUserByRole(session, r, spaceID, uidOrUsername, byUsername)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	segID := d.Get("isolation_segment").(string)
	if segID != "" && d.IsNewResource() {
		_, _, err := session.ClientV3.UpdateSpaceIsolationSegmentRelationship(spaceID, segID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if !d.IsNewResource() && d.HasChange("isolation_segment") {
		_, _, err := session.ClientV3.UpdateSpaceIsolationSegmentRelationship(spaceID, segID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = metadataUpdate(spaceMetadata, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	if !session.Config.DeleteRecursiveAllowed {
		// Check for apps
		apps, _, err := session.ClientV3.GetApplications(ccv3.Query{
			Key:    ccv3.SpaceGUIDFilter,
			Values: []string{d.Id()},
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if len(apps) > 0 {
			return diag.Errorf("Space %s has %d apps. Please delete them first or set delete_recursive_allowed to true", d.Id(), len(apps))
		}
		// Check routes
		routes, _, err := session.ClientV3.GetRoutes(ccv3.Query{
			Key:    ccv3.SpaceGUIDFilter,
			Values: []string{d.Id()},
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if len(routes) > 0 {
			return diag.Errorf("Space %s has %d routes. Please delete them first or set delete_recursive_allowed to true", d.Id(), len(routes))
		}
		// Check service instances
		serviceInstances, _, _, err := session.ClientV3.GetServiceInstances(ccv3.Query{
			Key:    ccv3.SpaceGUIDFilter,
			Values: []string{d.Id()},
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if len(serviceInstances) > 0 {
			return diag.Errorf("Space %s has %d service instances. Please delete them first or set delete_recursive_allowed to true", d.Id(), len(serviceInstances))
		}
	}

	j, _, err := session.ClientV3.DeleteSpace(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = session.ClientV3.PollJob(j)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(err)
}
