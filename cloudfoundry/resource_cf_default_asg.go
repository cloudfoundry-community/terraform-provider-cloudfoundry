package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDefaultAsg() *schema.Resource {

	return &schema.Resource{

		Create: resourceDefaultAsgCreate,
		Read:   resourceDefaultAsgRead,
		Update: resourceDefaultAsgUpdate,
		Delete: resourceDefaultAsgDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return ImportRead(resourceDefaultAsgRead)(d, meta)
			},
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDefaultRunningStagingName,
			},
			"asgs": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
		},
	}
}

func resourceDefaultAsgCreate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	name := d.Get("name").(string)
	asgs := d.Get("asgs").(*schema.Set).List()

	am := session.ClientV2
	switch name {
	case AppStatusRunning:
		for _, g := range asgs {
			_, err := am.BindRunningSecurityGroup(g.(string))
			if err != nil {
				return err
			}
		}
	case AppStatusStaging:
		for _, g := range asgs {
			_, err := am.BindStagingSecurityGroup(g.(string))
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("default security group name must be one of 'running' or 'staging'")
	}
	d.SetId(name)

	return nil
}

func resourceDefaultAsgRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	var asgs []ccv2.SecurityGroup
	var err error
	am := session.ClientV2
	tfAsgs := d.Get("asgs").(*schema.Set).List()
	switch d.Get("name").(string) {
	case AppStatusRunning:
		asgs, _, err = am.GetRunningSecurityGroups()
		if err != nil {
			return err
		}
	case AppStatusStaging:
		asgs, _, err = am.GetStagingSecurityGroups()
		if err != nil {
			return err
		}
	}

	finalTfAsgs := intersectSlices(tfAsgs, asgs, func(src, item interface{}) bool {
		return src.(string) == item.(ccv2.SecurityGroup).GUID
	})
	if IsImportState(d) && len(finalTfAsgs) == 0 {
		for _, asg := range asgs {
			finalTfAsgs = append(finalTfAsgs, asg.GUID)
		}
	}
	d.Set("asgs", schema.NewSet(resourceStringHash, finalTfAsgs))
	return nil
}

func resourceDefaultAsgUpdate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)

	secGroupToDelete, secGroupToAdd := getListChanges(d.GetChange("asgs"))
	am := session.ClientV2
	switch d.Get("name").(string) {
	case AppStatusRunning:
		for _, secGroup := range secGroupToAdd {
			_, err := am.BindRunningSecurityGroup(secGroup)
			if err != nil {
				return err
			}
		}
		for _, secGroup := range secGroupToDelete {
			_, err := am.UnbindRunningSecurityGroup(secGroup)
			if err != nil {
				return err
			}
		}
	case AppStatusStaging:
		for _, secGroup := range secGroupToAdd {
			_, err := am.BindStagingSecurityGroup(secGroup)
			if err != nil {
				return err
			}
		}
		for _, secGroup := range secGroupToDelete {
			_, err := am.UnbindStagingSecurityGroup(secGroup)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func resourceDefaultAsgDelete(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*managers.Session)

	am := session.ClientV2
	tfAsgs := d.Get("asgs").(*schema.Set).List()

	switch d.Get("name").(string) {
	case AppStatusRunning:
		for _, asg := range tfAsgs {
			_, err := am.UnbindRunningSecurityGroup(asg.(string))
			if err != nil {
				return err
			}
		}
	case AppStatusStaging:
		for _, asg := range tfAsgs {
			_, err := am.UnbindStagingSecurityGroup(asg.(string))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
