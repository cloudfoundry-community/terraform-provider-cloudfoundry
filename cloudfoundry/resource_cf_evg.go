package cloudfoundry

import (
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				return ImportRead(resourceEvgRead)(d, meta)
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

func resourceEvgRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	var variables map[string]string
	var err error
	switch d.Get("name").(string) {
	case AppStatusRunning:
		variables, _, err = session.ClientV2.GetEnvVarGroupRunning()
	case AppStatusStaging:
		variables, _, err = session.ClientV2.GetEnvVarGroupStaging()
	}
	if err != nil {
		return err
	}
	finalVariables := make(map[string]interface{})
	tfVariables := d.Get("variables").(map[string]interface{})
	for tfKey := range tfVariables {
		if v, ok := variables[tfKey]; ok {
			finalVariables[tfKey] = v
		}
	}

	if IsImportState(d) && len(finalVariables) == 0 {
		for k, v := range variables {
			finalVariables[k] = v
		}
	}
	d.Set("variables", finalVariables)
	return nil
}

func resourceEvgUpdate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	tfVariables := d.Get("variables").(map[string]interface{})

	var variables map[string]string
	var err error
	switch name {
	case AppStatusRunning:
		variables, _, err = session.ClientV2.GetEnvVarGroupRunning()
	case AppStatusStaging:
		variables, _, err = session.ClientV2.GetEnvVarGroupStaging()
	}
	if err != nil {
		return err
	}
	old, new := d.GetChange("variables")
	keyToDelete, keyToAdd := getMapChanges(old, new)
	for _, key := range keyToAdd {
		variables[key] = tfVariables[key].(string)
	}
	for _, key := range keyToDelete {
		delete(variables, key)
	}

	switch name {
	case AppStatusRunning:
		_, err = session.ClientV2.SetEnvVarGroupRunning(variables)
	case AppStatusStaging:
		_, err = session.ClientV2.SetEnvVarGroupStaging(variables)
	}
	if err != nil {
		return err
	}
	return nil
}

func resourceEvgDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	var variables map[string]string
	var err error
	switch d.Get("name").(string) {
	case AppStatusRunning:
		variables, _, err = session.ClientV2.GetEnvVarGroupRunning()
	case AppStatusStaging:
		variables, _, err = session.ClientV2.GetEnvVarGroupStaging()
	}
	if err != nil {
		return err
	}
	for k := range d.Get("variables").(map[string]interface{}) {
		delete(variables, k)
	}
	switch d.Get("name").(string) {
	case AppStatusRunning:
		_, err = session.ClientV2.SetEnvVarGroupRunning(variables)
	case AppStatusStaging:
		_, err = session.ClientV2.SetEnvVarGroupStaging(variables)
	}
	if err != nil {
		return err
	}
	return nil
}
