package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
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
			"asgs": &schema.Schema{
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
	asgs := d.Get("asgs").(*schema.Set).List()

	am := session.ASGManager()
	switch name {
	case AppStatusRunning:
		if err = am.UnbindAllFromRunning(); err != nil {
			return err
		}
		for _, g := range asgs {
			if err = am.BindToRunning(g.(string)); err != nil {
				return err
			}
		}
	case AppStatusStaging:
		if err = am.UnbindAllFromStaging(); err != nil {
			return err
		}
		for _, g := range asgs {
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

	var asgs []string

	am := session.ASGManager()
	switch d.Get("name").(string) {
	case AppStatusRunning:
		if asgs, err = am.Running(); err != nil {
			return err
		}
	case AppStatusStaging:
		if asgs, err = am.Staging(); err != nil {
			return err
		}
	}

	tfAsgs := []interface{}{}
	for _, s := range asgs {
		tfAsgs = append(tfAsgs, s)
	}
	d.Set("asgs", schema.NewSet(resourceStringHash, tfAsgs))
	return nil
}

func resourceDefaultAsgUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var asgs []string

	tfAsgs := d.Get("asgs").(*schema.Set).List()

	am := session.ASGManager()
	switch d.Get("name").(string) {
	case AppStatusRunning:
		if asgs, err = am.Running(); err != nil {
			return err
		}
		for _, s := range tfAsgs {
			asg := s.(string)
			if !isStringInList(asgs, asg) {
				if err = am.BindToRunning(asg); err != nil {
					return err
				}
			}
		}
		for _, s := range asgs {
			if !isStringInInterfaceList(tfAsgs, s) {
				if err = am.UnbindFromRunning(s); err != nil {
					return err
				}
			}
		}
	case AppStatusStaging:
		if asgs, err = am.Staging(); err != nil {
			return err
		}
		for _, s := range tfAsgs {
			asg := s.(string)
			if !isStringInList(asgs, asg) {
				err = am.BindToStaging(asg)
				if err != nil {
					return err
				}
			}
		}
		for _, s := range asgs {
			if !isStringInInterfaceList(tfAsgs, s) {
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
