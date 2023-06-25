package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceSpaceAsgs() *schema.Resource {

	return &schema.Resource{

		CreateContext: resourceSpaceAsgsCreate,
		ReadContext:   resourceSpaceAsgsRead,
		UpdateContext: resourceSpaceAsgsUpdate,
		DeleteContext: resourceSpaceAsgsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: ImportReadContext(resourceSpaceUsersRead),
		},

		Schema: map[string]*schema.Schema{
			"space": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"running_asgs": {
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        resourceStringHash,
			},
			"staging_asgs": {
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

func resourceSpaceAsgsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	return resourceSpaceAsgsUpdate(ctx, d, meta)
}

func resourceSpaceAsgsRead(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	spaceId := d.Get("space").(string)

	runningAsgs, _, err := session.ClientV2.GetSpaceSecurityGroups(spaceId)
	if err != nil {
		return nil
	}
	if !IsImportState(d) {
		finalRunningAsg := intersectSlices(d.Get("running_asgs").(*schema.Set).List(), runningAsgs, func(source, item interface{}) bool {
			return source.(string) == item.(ccv2.SecurityGroup).GUID
		})
		d.Set("running_asgs", schema.NewSet(resourceStringHash, finalRunningAsg))
	} else {
		finalRunningAsgs, _ := getInSlice(runningAsgs, func(object interface{}) bool {
			return !object.(ccv2.SecurityGroup).RunningDefault
		})
		d.Set("running_asgs", schema.NewSet(resourceStringHash, objectsToIds(finalRunningAsgs, func(object interface{}) string {
			return object.(ccv2.SecurityGroup).GUID
		})))
	}

	stagingAsgs, _, err := session.ClientV2.GetSpaceStagingSecurityGroups(spaceId)
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

	return nil
}

func resourceSpaceAsgsUpdate(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	spaceID := d.Get("space").(string)
	_, _, err := session.ClientV2.GetSpace(spaceID)
	if err != nil {
		return diag.FromErr(err)
	}

	removeRunningAsgs, addRunningAsgs := getListChanges(d.GetChange("running_asgs"))
	for _, asgID := range removeRunningAsgs {
		err := updateSpaceRemoveRunningAsgs(c, spaceID, asgID, session)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	for _, asgID := range addRunningAsgs {
		err := updateSpaceAddRunningAsgs(c, spaceID, asgID, session)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	removeStagingAsgs, addStagingAsgs := getListChanges(d.GetChange("staging_asgs"))
	for _, asgID := range removeStagingAsgs {
		err := updateSpaceRemoveStagingAsgs(c, spaceID, asgID, session)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	for _, asgID := range addStagingAsgs {
		err := updateSpaceAddStagingAsgs(c, spaceID, asgID, session)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceSpaceAsgsDelete(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	spaceID := d.Get("space").(string)
	_, _, err := session.ClientV2.GetSpace(spaceID)
	if err != nil {
		return diag.FromErr(err)
	}

	removeRunningAsgs := d.Get("running_asgs").(*schema.Set).List()
	for _, asgID := range removeRunningAsgs {
		err := updateSpaceRemoveRunningAsgs(c, spaceID, asgID.(string), session)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	removeStagingAsgs := d.Get("staging_asgs").(*schema.Set).List()
	for _, asgID := range removeStagingAsgs {
		err := updateSpaceRemoveStagingAsgs(c, spaceID, asgID.(string), session)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

// helper functions
// These should go to the ClientV2; would someone like to raise a change request?
func updateSpaceAddRunningAsgs(c context.Context, spaceID string, asgID string, session *managers.Session) error {
	return updateSpaceAsgs(c, "PUT", "/v2/spaces/%s/security_groups/%s", spaceID, asgID, session)
}

func updateSpaceRemoveRunningAsgs(c context.Context, spaceID string, asgID string, session *managers.Session) error {
	return updateSpaceAsgs(c, "DELETE", "/v2/spaces/%s/security_groups/%s", spaceID, asgID, session)
}

func updateSpaceAddStagingAsgs(c context.Context, spaceID string, asgID string, session *managers.Session) error {
	return updateSpaceAsgs(c, "PUT", "/v2/spaces/%s/staging_security_groups/%s", spaceID, asgID, session)
}

func updateSpaceRemoveStagingAsgs(c context.Context, spaceID string, asgID string, session *managers.Session) error {
	return updateSpaceAsgs(c, "DELETE", "/v2/spaces/%s/staging_security_groups/%s", spaceID, asgID, session)
}

func updateSpaceAsgs(c context.Context, method string, path string, spaceID string, asgID string, session *managers.Session) error {
	p := fmt.Sprintf(path, spaceID, asgID)

	req, err := session.RawClient.NewRequest(method, p, nil)
	if err != nil {
		return err
	}

	resp, err := session.RawClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		err = ccerror.RawHTTPStatusError{
			StatusCode: resp.StatusCode,
		}
	}
	err = resp.Body.Close()
	return err
}
