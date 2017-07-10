package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

func resourceConfig() *schema.Resource {

	return &schema.Resource{

		Create: resourceConfigCreate,
		Read:   resourceConfigRead,
		Update: resourceConfigUpdate,
		Delete: resourceConfigDelete,

		Schema: map[string]*schema.Schema{
			"feature_flags": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_org_creation": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"private_domain_creation": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"app_bits_upload": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"app_scaling": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"route_creation": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"service_instance_creation": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"diego_docker": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"set_roles_by_username": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"unset_roles_by_username": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"task_creation": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"env_var_visibility": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"space_scoped_private_broker_creation": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"space_developer_env_var_visibility": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
					},
				},
			},
		},
	}
}

func validateFeatureFlagValue(v interface{}, k string) (ws []string, errs []error) {
	value := v.(string)
	if value != "enabled" && value != "disabled" {
		errs = append(errs, fmt.Errorf("%q must be one of 'enabled' or 'disabled'", k))
	}
	return
}

func resourceConfigCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	if v, ok := d.GetOk("feature_flags"); ok {
		if err = session.SetFeatureFlags(getFeatureFlags(v)); err != nil {
			return
		}
	}

	d.SetId("config")
	return resourceConfigRead(d, meta)
}

func resourceConfigRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var featureFlags map[string]bool
	if featureFlags, err = session.GetFeatureFlags(); err != nil {
		return
	}

	flags := make(map[string]interface{})
	for k, v := range featureFlags {
		if v {
			flags[k] = "enabled"
		} else {
			flags[k] = "disabled"
		}
	}

	d.Set("feature_flags", []interface{}{flags})
	return
}

func resourceConfigUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	if d.HasChange("feature_flags") {
		if err = session.SetFeatureFlags(getFeatureFlags(d.Get("feature_flags"))); err != nil {
			return
		}
	}
	return
}

func resourceConfigDelete(d *schema.ResourceData, meta interface{}) (err error) {
	return
}

func getFeatureFlags(v interface{}) map[string]bool {

	featureFlags := make(map[string]bool)
	for k, v := range v.([]interface{})[0].(map[string]interface{}) {

		vv := v.(string)
		if len(vv) > 0 {
			featureFlags[k] = (vv == "enabled")
		}
	}
	return featureFlags
}
