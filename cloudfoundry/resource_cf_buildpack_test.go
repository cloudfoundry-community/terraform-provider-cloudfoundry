package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
	"os"
)

const buildpackResource = `

variable "tomee_buildpack_ver" {
	default="v4.1"
}

variable "github_user" {}
variable "github_password" {}

resource "cloudfoundry_buildpack" "tomee" {
	
	name = "tomee-buildpack"

	github_release {
		owner = "cloudfoundry-community"
		repo = "tomee-buildpack"
		version = "${var.tomee_buildpack_ver}"
		filename = "tomee-buildpack-v4.1.zip"
		user = "%s"
		password = "%s"
	}
}
`

const buildpackResourceUpdate1 = `

variable "tomee_buildpack_ver" {
	default="v4.1"
}

resource "cloudfoundry_buildpack" "tomee" {
	
	name = "tomee-buildpack"
	position = 5
	enabled = false
	locked = true

	github_release {
		owner = "cloudfoundry-community"
		repo = "tomee-buildpack"
		version = "${var.tomee_buildpack_ver}"
		filename = "tomee-buildpack-v4.1.zip"
		user = "%s"
		password = "%s"
	}
}
`

const buildpackResourceUpdate2 = `

resource "cloudfoundry_buildpack" "tomee" {
	
	name = "tomee-buildpack"
	position = 5
	enabled = true
	locked = false

	git {
		url = "https://github.com/cloudfoundry-community/tomee-buildpack"				
	}
}
`

func TestAccBuildpack_normal(t *testing.T) {

	refBuildpack := "cloudfoundry_buildpack.tomee"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckBuildpackDestroyed("tomee-buildpack"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(buildpackResource, os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_PASSWORD")),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckBuildpackExists(refBuildpack, "tomee-buildpack-v4.1.zip"),
						resource.TestCheckResourceAttr(
							refBuildpack, "name", "tomee-buildpack"),
						resource.TestCheckResourceAttr(
							refBuildpack, "position", "1"),
						resource.TestCheckResourceAttr(
							refBuildpack, "enabled", "true"),
						resource.TestCheckResourceAttr(
							refBuildpack, "locked", "false"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(buildpackResourceUpdate1, os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_PASSWORD")),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckBuildpackExists(refBuildpack, "tomee-buildpack-v4.1.zip"),
						resource.TestCheckResourceAttr(
							refBuildpack, "name", "tomee-buildpack"),
						resource.TestCheckResourceAttr(
							refBuildpack, "position", "5"),
						resource.TestCheckResourceAttr(
							refBuildpack, "enabled", "false"),
						resource.TestCheckResourceAttr(
							refBuildpack, "locked", "true"),
					),
				},

				resource.TestStep{
					Config: buildpackResourceUpdate2,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckBuildpackExists(refBuildpack, "tomee-buildpack.zip"),
						resource.TestCheckResourceAttr(
							refBuildpack, "name", "tomee-buildpack"),
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

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[refBuildpack]
		if !ok {
			return fmt.Errorf("buildpack resource '%s' not found in terraform state", refBuildpack)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			refBuildpack, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var bp cfapi.CCBuildpack
		bpm := session.BuildpackManager()
		if bp, err = bpm.ReadBuildpack(id); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved buildpack for resource '%s' with id '%s': %# v",
			refBuildpack, id, bp)

		if err := assertEquals(attributes, "name", bp.Name); err != nil {
			return err
		}
		if err := assertEquals(attributes, "position", bp.Position); err != nil {
			return err
		}
		if err := assertEquals(attributes, "locked", bp.Locked); err != nil {
			return err
		}
		if err := assertEquals(attributes, "enabled", bp.Enabled); err != nil {
			return err
		}
		if bp.Filename != bpFilename {
			return fmt.Errorf("expected buildpack file name to be 'tomee-buildpack-v4.1.zip' but it was '%s'", bp.Filename)
		}
		return
	}
}

func testAccCheckBuildpackDestroyed(buildpackName string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)
		if _, err := session.BuildpackManager().FindBuildpack(buildpackName); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}
		return fmt.Errorf("buildpack with name '%s' still exists in cloud foundry", buildpackName)
	}
}
