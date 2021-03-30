package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const orgDataResource = `

resource "cloudfoundry_org" "myorg" {
	name = "myorg-ds-org"
}

data "cloudfoundry_org" "dd" {
    name = "${cloudfoundry_org.myorg.name}"
}
`

func TestAccDataSourceOrg_normal(t *testing.T) {

	ref := "data.cloudfoundry_org.dd"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: orgDataResource,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceOrgExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "myorg-ds-org"),
					),
				},
			},
		})
}

func checkDataSourceOrgExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("org '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]

		orgs, _, err := session.ClientV2.GetOrganizations(ccv2.FilterByName(name))
		if err != nil {
			return err
		}
		if len(orgs) == 0 {
			return NotFound
		}
		if err := assertSame(id, orgs[0].GUID); err != nil {
			return err
		}

		return nil
	}
}
