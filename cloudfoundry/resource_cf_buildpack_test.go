package cloudfoundry

import (
	"fmt"
	"path/filepath"
	"testing"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const buildpackResource = `
resource "cloudfoundry_buildpack" "tomee" {

	name = "tomee-buildpack-res"

	path = "https://github.com/cloudfoundry-community/tomee-buildpack/releases/download/v4.3/tomee-buildpack-v4.3.zip"
}
`

const buildpackResourceUpdate1 = `

resource "cloudfoundry_buildpack" "tomee" {

	name = "tomee-buildpack-res"
	position = 5
	enabled = false
	locked = true

	path = "https://github.com/cloudfoundry-community/tomee-buildpack/releases/download/v4.3/tomee-buildpack-v4.3.zip"
}
`

const buildpackResourceUpdate2 = `

resource "cloudfoundry_buildpack" "tomee" {

	name = "tomee-buildpack-res"
	position = 5
	enabled = true
	locked = false

	path = "%s"
}
`

func TestAccResBuildpack_normal(t *testing.T) {

	fixturesBp := asset("buildpacks")
	refBuildpack := "cloudfoundry_buildpack.tomee"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckBuildpackDestroyed("tomee-buildpack-res"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: buildpackResource,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckBuildpackExists(refBuildpack, "tomee-buildpack-v4.3.zip"),
						resource.TestCheckResourceAttr(
							refBuildpack, "name", "tomee-buildpack-res"),
						resource.TestCheckResourceAttr(
							refBuildpack, "position", "1"),
						resource.TestCheckResourceAttr(
							refBuildpack, "enabled", "true"),
						resource.TestCheckResourceAttr(
							refBuildpack, "locked", "false"),
					),
				},
				resource.TestStep{
					ResourceName:            refBuildpack,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"path"},
				},
				resource.TestStep{
					Config: buildpackResourceUpdate1,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckBuildpackExists(refBuildpack, "tomee-buildpack-v4.3.zip"),
						resource.TestCheckResourceAttr(
							refBuildpack, "name", "tomee-buildpack-res"),
						resource.TestCheckResourceAttr(
							refBuildpack, "position", "5"),
						resource.TestCheckResourceAttr(
							refBuildpack, "enabled", "false"),
						resource.TestCheckResourceAttr(
							refBuildpack, "locked", "true"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(buildpackResourceUpdate2, filepath.Join(fixturesBp, "tomee-buildpack-v4.5.2.zip")),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckBuildpackExists(refBuildpack, "tomee-buildpack-v4.5.2.zip"),
						resource.TestCheckResourceAttr(
							refBuildpack, "name", "tomee-buildpack-res"),
						resource.TestCheckResourceAttr(
							refBuildpack, "position", "5"),
						resource.TestCheckResourceAttr(
							refBuildpack, "enabled", "true"),
						resource.TestCheckResourceAttr(
							refBuildpack, "locked", "false"),
					),
				},
			},
		})
}

func testAccCheckBuildpackExists(refBuildpack, bpFilename string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[refBuildpack]
		if !ok {
			return fmt.Errorf("buildpack resource '%s' not found in terraform state", refBuildpack)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		bp, _, err := session.ClientV2.GetBuildpack(id)
		if err != nil {
			return err
		}

		if err := assertEquals(attributes, "name", bp.Name); err != nil {
			return err
		}
		if err := assertEquals(attributes, "position", bp.Position.Value); err != nil {
			return err
		}
		if err := assertEquals(attributes, "locked", bp.Locked.Value); err != nil {
			return err
		}
		if err := assertEquals(attributes, "enabled", bp.Enabled.Value); err != nil {
			return err
		}
		if bp.Filename != bpFilename {
			return fmt.Errorf("expected buildpack file name to be '%s' but it was '%s'", bpFilename, bp.Filename)
		}
		return nil
	}
}

func testAccCheckBuildpackDestroyed(buildpackName string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)
		bps, _, err := session.ClientV2.GetBuildpacks(ccv2.FilterByName(buildpackName))
		if err != nil {
			return err
		}
		if len(bps) > 0 {
			return fmt.Errorf("buildpack with name '%s' still exists in cloud foundry", buildpackName)
		}
		return nil
	}
}
