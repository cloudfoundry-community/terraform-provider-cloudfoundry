package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FlagStatusEnabled - Status returned by CF api for enabled flags
const FlagStatusEnabled = "enabled"

// FlagStatusDisabled - Status returned by CF api for disabled flags
const FlagStatusDisabled = "disabled"

func resourceConfig() *schema.Resource {

	return &schema.Resource{

		Create: resourceConfigCreate,
		Read:   resourceConfigRead,
		Update: resourceConfigUpdate,
		Delete: resourceConfigDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.SetId("config")
				return ImportRead(resourceConfigRead)(d, meta)
			},
		},

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
						"service_instance_sharing": &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateFeatureFlagValue,
							Optional:     true,
							Computed:     true,
						},
						"hide_marketplace_from_unauthenticated_users": &schema.Schema{
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
	if value != FlagStatusEnabled && value != FlagStatusDisabled {
		errs = append(errs, fmt.Errorf("%q must be one of 'enabled' or 'disabled'", k))
	}
	return ws, errs
}

func resourceConfigCreate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	if v, ok := d.GetOk("feature_flags"); ok {
		ffs := getFeatureFlags(v)
		for _, ff := range ffs {
			_, err := session.ClientV2.SetConfigFeatureFlags(ff)
			if err != nil {
				return err
			}
		}

	}

	d.SetId("config")
	return resourceConfigRead(d, meta)
}

func resourceConfigRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	featureFlags, _, err := session.ClientV2.GetConfigFeatureFlags()
	if err != nil {
		return err
	}

	flags := make(map[string]interface{})
	for _, ff := range featureFlags {
		if ff.Enabled {
			flags[ff.Name] = FlagStatusEnabled
		} else {
			flags[ff.Name] = FlagStatusDisabled
		}
	}

	d.Set("feature_flags", []interface{}{flags})
	return err
}

func resourceConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	if d.HasChange("feature_flags") {
		ffs := getFeatureFlags(d.Get("feature_flags"))
		for _, ff := range ffs {
			_, err := session.ClientV2.SetConfigFeatureFlags(ff)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func resourceConfigDelete(d *schema.ResourceData, meta interface{}) (err error) {
	return nil
}

func getFeatureFlags(v interface{}) []ccv2.FeatureFlag {
	flags := v.([]interface{})[0].(map[string]interface{})
	featureFlags := make([]ccv2.FeatureFlag, 0)
	for k, v := range flags {

		vv := v.(string)
		if len(vv) > 0 {
			featureFlags = append(featureFlags, ccv2.FeatureFlag{
				Name:    k,
				Enabled: vv == FlagStatusEnabled,
			})
		}
	}
	return featureFlags
}
