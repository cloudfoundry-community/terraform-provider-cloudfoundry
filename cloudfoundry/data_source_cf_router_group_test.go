package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const routerGroupDataResource = `

data "cloudfoundry_router_group" "rg" {
    name = "default-tcp"
}
`

func TestAccDataSourceRouterGroup_normal(t *testing.T) {

	ref := "data.cloudfoundry_router_group.rg"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: routerGroupDataResource,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceRouterGroupExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "default-tcp"),
					),
				},
			},
		})
}

func checkDataSourceRouterGroupExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("router_group '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]
		rgType := rs.Primary.Attributes["type"]

		routerGroup, err := session.RouterClient.GetRouterGroupByName(name)
		if err != nil {
			return err
		}
		if err := assertSame(id, routerGroup.GUID); err != nil {
			return err
		}
		if err := assertSame(rgType, routerGroup.Type); err != nil {
			return err
		}

		return nil
	}
}
