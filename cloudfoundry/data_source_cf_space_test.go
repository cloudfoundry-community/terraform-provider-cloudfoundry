package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

const spaceDataResource1 = `

resource "cloudfoundry_org" "org1" {
	name = "organization-one"
}
resource "cloudfoundry_space" "space1" {
	name = "space-one"
	org = "${cloudfoundry_org.org1.id}"
}

data "cloudfoundry_space" "myspace" {
    name = "${cloudfoundry_space.space1.name}"
	org = "${cloudfoundry_org.org1.id}"
}
`

const spaceDataResource2 = `

data "cloudfoundry_space" "default" {
    name = "pcfdev-space"
	org_name = "pcfdev-org"
}
`

func TestAccDataSourceSpace_normal(t *testing.T) {

	ref1 := "data.cloudfoundry_space.myspace"
	ref2 := "data.cloudfoundry_space.default"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: spaceDataResource1,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceSpaceExists(ref1),
						resource.TestCheckResourceAttr(
							ref1, "name", "space-one"),
						resource.TestCheckResourceAttr(
							ref1, "org_name", "organization-one"),
						resource.TestCheckResourceAttrSet(
							ref1, "org"),
					),
				},

				resource.TestStep{
					Config: spaceDataResource2,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceSpaceExists(ref2),
						resource.TestCheckResourceAttr(
							ref2, "name", "pcfdev-space"),
						resource.TestCheckResourceAttr(
							ref2, "org_name", "pcfdev-org"),
						resource.TestCheckResourceAttr(
							ref2, "org", defaultPcfDevOrgID()),
					),
				},
			},
		})
}

func checkDataSourceSpaceExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("space '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]
		org := rs.Primary.Attributes["org"]

		var (
			err   error
			space cfapi.CCSpace
		)

		space, err = session.SpaceManager().FindSpaceInOrg(name, org)
		if err != nil {
			return err
		}
		if err := assertSame(id, space.ID); err != nil {
			return err
		}

		return nil
	}
}
