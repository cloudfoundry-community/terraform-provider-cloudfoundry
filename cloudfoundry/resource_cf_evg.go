package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceEvg() *schema.Resource {

	return &schema.Resource{

		Create: resourceEvgCreate,
		Read:   resourceEvgRead,
		Update: resourceEvgUpdate,
		Delete: resourceEvgDelete,

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
				ValidateFunc: validateDefaultRunningStagingName,
			},
			"variables": &schema.Schema{
				Type:     schema.TypeMap,
				Required: true,
			},
		},
	}
}

func resourceEvgCreate(d *schema.ResourceData, meta interface{}) (err error) {

	if err = resourceEvgUpdate(d, meta); err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceEvgRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var variables map[string]interface{}
	if variables, err = session.EVGManager().GetEVG(d.Get("name").(string)); err != nil {
		return
	}
	d.Set("variables", variables)
	return nil
}

func resourceEvgUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	name := d.Get("name").(string)
	variables := d.Get("variables").(map[string]interface{})

	return session.EVGManager().SetEVG(name, variables)
}

func resourceEvgDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	return session.EVGManager().SetEVG(d.Get("name").(string), map[string]interface{}{})
}
