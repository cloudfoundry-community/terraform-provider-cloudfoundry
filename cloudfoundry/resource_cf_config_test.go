package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const configResource = `
resource "cf_config" "ff" {
	feature_flags{
		route_creation = "disabled"
		task_creation = "enabled"
		env_var_visibility = "disabled"
	}
}
`

const configResourceUpdate = `
resource "cf_config" "ff" {
	feature_flags{
		route_creation = "enabled"
		task_creation = "disabled"
		env_var_visibility = "enabled"
	}
}
`

func TestAccConfig_normal(t *testing.T) {

	resConfig := "cf_config.ff"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckConfigDestroy,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: configResource,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckConfig(resConfig),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.user_org_creation", "disabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.private_domain_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.app_bits_upload", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.app_scaling", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.route_creation", "disabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.service_instance_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.diego_docker", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.set_roles_by_username", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.unset_roles_by_username", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.task_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.env_var_visibility", "disabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.space_scoped_private_broker_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.space_developer_env_var_visibility", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.service_instance_sharing", "disabled"),
					),
				},

				resource.TestStep{
					Config: configResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckConfig(resConfig),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.user_org_creation", "disabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.private_domain_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.app_bits_upload", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.app_scaling", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.route_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.service_instance_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.diego_docker", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.set_roles_by_username", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.unset_roles_by_username", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.task_creation", "disabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.env_var_visibility", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.space_scoped_private_broker_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.space_developer_env_var_visibility", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.service_instance_sharing", "disabled"),
					),
				},
			},
		})
}

func testAccCheckConfig(resConfig string) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resConfig]
		if !ok {
			return fmt.Errorf("'%s' resource not found in terraform state", resConfig)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v", resConfig, rs)

		attributes := rs.Primary.Attributes

		var featureFlags map[string]bool
		if featureFlags, err = session.GetFeatureFlags(); err != nil {
			return
		}

		session.Log.DebugMessage(
			"retrieved feature flags: %# v", featureFlags)

		if err := assertListEquals(attributes, "feature_flags", 1,
			func(values map[string]string, i int) (match bool) {

				if len(values) == len(featureFlags) {

					var (
						vv interface{}
						ok bool
					)

					for k, v := range values {

						if vv, ok = featureFlags[k]; ok {
							if vv.(bool) {
								ok = (v == "enabled")
							} else {
								ok = (v == "disabled")
							}
						}
						if !ok {
							return false
						}
					}
					return true
				}

				return false

			}); err != nil {
			return err
		}

		return
	}
}

func testAccCheckConfigDestroy(s *terraform.State) error {
	return nil
}
