package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const spaceDataResource = `

resource "cf_org" "org1" {
	name = "organization-one"
}
resource "cf_space" "space1" {
	name = "space-one"
	org = "${cf_org.org1.id}"
}

data "cf_space" "myspace" {
    name = "${cf_space.space1.name}"
	org = "${cf_org.org1.id}"
}
`

func TestAccDataSourceSpace_normal(t *testing.T) {

	ref := "data.cf_space.myspace"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: spaceDataResource,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceSpaceExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "space-one"),
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
