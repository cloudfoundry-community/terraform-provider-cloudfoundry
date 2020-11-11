package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const spaceDataResource1 = `

resource "cloudfoundry_org" "org1" {
	name = "organization-ds-space"
}
resource "cloudfoundry_space" "space1" {
	name = "space-ds-space"
	org = "${cloudfoundry_org.org1.id}"
}

data "cloudfoundry_space" "myspace" {
	name = "${cloudfoundry_space.space1.name}"
	org = "${cloudfoundry_org.org1.id}"
}
`

const spaceDataResource2 = `

data "cloudfoundry_space" "default" {
	name = "%s"
	org_name = "%s"
}
`

func TestAccDataSourceSpace_normal(t *testing.T) {

	ref1 := "data.cloudfoundry_space.myspace"
	ref2 := "data.cloudfoundry_space.default"

	orgID, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: spaceDataResource1,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceSpaceExists(ref1),
						resource.TestCheckResourceAttr(
							ref1, "name", "space-ds-space"),
						resource.TestCheckResourceAttr(
							ref1, "org_name", "organization-ds-space"),
						resource.TestCheckResourceAttrSet(
							ref1, "org"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(spaceDataResource2, spaceName, orgName),
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceSpaceExists(ref2),
						resource.TestCheckResourceAttr(
							ref2, "name", spaceName),
						resource.TestCheckResourceAttr(
							ref2, "org_name", orgName),
						resource.TestCheckResourceAttr(
							ref2, "org", orgID),
					),
				},
			},
		})
}

func checkDataSourceSpaceExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("space '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]
		org := rs.Primary.Attributes["org"]

		spaces, _, err := session.ClientV2.GetSpaces(ccv2.FilterByName(name), ccv2.FilterByOrg(org))
		if err != nil {
			return err
		}
		if len(spaces) == 0 {
			return NotFound
		}
		if err := assertSame(id, spaces[0].GUID); err != nil {
			return err
		}

		return nil
	}
}
