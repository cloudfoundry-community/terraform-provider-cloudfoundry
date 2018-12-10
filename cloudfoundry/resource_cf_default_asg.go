package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
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
				return ImportStatePassthrough(d, meta)
			},
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDefaultRunningStagingName,
			},
			"asg_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
			},
		},
	}
}

func resourceDefaultAsgCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	name := d.Get("name").(string)
	asgIDs := d.Get("asg_ids").(*schema.Set).List()

	am := session.ASGManager()
	switch name {
	case AppStatusRunning:
		if err = am.UnbindAllFromRunning(); err != nil {
			return err
		}
		for _, g := range asgIDs {
			if err = am.BindToRunning(g.(string)); err != nil {
				return err
			}
		}
	case AppStatusStaging:
		if err = am.UnbindAllFromStaging(); err != nil {
			return err
		}
		for _, g := range asgIDs {
			if err = am.BindToStaging(g.(string)); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("default security group name must be one of 'running' or 'staging'")
	}
	d.SetId(name)

	return nil
}

func resourceDefaultAsgRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var asgIDs []string

	am := session.ASGManager()
	switch d.Get("name").(string) {
	case AppStatusRunning:
		if asgIDs, err = am.Running(); err != nil {
			return err
		}
	case AppStatusStaging:
		if asgIDs, err = am.Staging(); err != nil {
			return err
		}
	}

	tfAsgIDs := []interface{}{}
	for _, s := range asgIDs {
		tfAsgIDs = append(tfAsgIDs, s)
	}
	d.Set("asg_ids", schema.NewSet(resourceStringHash, tfAsgIDs))
	return nil
}

func resourceDefaultAsgUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var asgIDs []string

	tfAsgIDs := d.Get("asg_ids").(*schema.Set).List()

	am := session.ASGManager()
	switch d.Get("name").(string) {
	case AppStatusRunning:
		if asgIDs, err = am.Running(); err != nil {
			return err
		}
		for _, s := range tfAsgIDs {
			asg := s.(string)
			if !isStringInList(asgIDs, asg) {
				if err = am.BindToRunning(asg); err != nil {
					return err
				}
			}
		}
		for _, s := range asgIDs {
			if !isStringInInterfaceList(tfAsgIDs, s) {
				if err = am.UnbindFromRunning(s); err != nil {
					return err
				}
			}
		}
	case AppStatusStaging:
		if asgIDs, err = am.Staging(); err != nil {
			return err
		}
		for _, s := range tfAsgIDs {
			asg := s.(string)
			if !isStringInList(asgIDs, asg) {
				err = am.BindToStaging(asg)
				if err != nil {
					return err
				}
			}
		}
		for _, s := range asgIDs {
			if !isStringInInterfaceList(tfAsgIDs, s) {
				if err = am.UnbindFromStaging(s); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func resourceDefaultAsgDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.ASGManager()
	switch d.Get("name").(string) {
	case AppStatusRunning:
		err = am.UnbindAllFromRunning()
		if err != nil {
			return err
		}
	case AppStatusStaging:
		err = am.UnbindAllFromStaging()
		if err != nil {
			return err
		}
	}
	return nil
}
