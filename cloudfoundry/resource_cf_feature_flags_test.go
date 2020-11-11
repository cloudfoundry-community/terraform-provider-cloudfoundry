package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const configResource = `
resource "cloudfoundry_feature_flags" "ff" {
	feature_flags {
		user_org_creation = "enabled"
		private_domain_creation = "disabled"
	}
}
`

const configResourceUpdate = `
resource "cloudfoundry_feature_flags" "ff" {
	feature_flags {
		user_org_creation = "disabled"
		private_domain_creation = "enabled"
	}
}
`

func TestAccResConfig_normal(t *testing.T) {

	resConfig := "cloudfoundry_feature_flags.ff"

	resource.ParallelTest(t,
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
							resConfig, "feature_flags.0.user_org_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.private_domain_creation", "disabled"),
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
							resConfig, "feature_flags.0.task_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.env_var_visibility", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.space_scoped_private_broker_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.space_developer_env_var_visibility", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.service_instance_sharing", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.hide_marketplace_from_unauthenticated_users", "disabled"),
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
							resConfig, "feature_flags.0.task_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.env_var_visibility", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.space_scoped_private_broker_creation", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.space_developer_env_var_visibility", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.service_instance_sharing", "enabled"),
						resource.TestCheckResourceAttr(
							resConfig, "feature_flags.0.hide_marketplace_from_unauthenticated_users", "disabled"),
					),
				},
				resource.TestStep{
					ResourceName:      resConfig,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}

func testAccCheckConfig(resConfig string) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resConfig]
		if !ok {
			return fmt.Errorf("'%s' resource not found in terraform state", resConfig)
		}

		attributes := rs.Primary.Attributes

		featureFlags, _, err := session.ClientV2.GetConfigFeatureFlags()
		if err != nil {
			return
		}
		if err := assertListEquals(attributes, "feature_flags", 1,
			func(values map[string]string, i int) (match bool) {

				if len(values) == len(featureFlags) {

					var (
						ok bool
					)

					for k, v := range values {
						for _, ff := range featureFlags {
							if ff.Name == k {
								if ff.Enabled {
									ok = (v == "enabled")
								} else {
									ok = (v == "disabled")
								}
								break
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
