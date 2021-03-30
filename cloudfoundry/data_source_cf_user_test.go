package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const userDataResource = `

data "cloudfoundry_user" "admin-user" {
    name = "admin"
}
`

func TestAccDataSourceUser_normal(t *testing.T) {

	ref := "data.cloudfoundry_user.admin-user"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: userDataResource,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "admin"),
					),
				},
			},
		})
}

func checkDataSourceUserExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("user '%s' not found in terraform state", resource)
		}
		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]

		users, err := session.ClientUAA.GetUsersByUsername(name)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			return NotFound
		}
		if err := assertSame(id, users[0].ID); err != nil {
			return err
		}

		return nil
	}
}
