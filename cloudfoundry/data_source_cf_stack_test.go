package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const stackDataResource = `

data "cloudfoundry_stack" "s" {
    name = "%s"
}
`

func TestAccDataSourceStack_normal(t *testing.T) {
	defaultStacks, _, err := testSession().ClientV2.GetStacks()
	if err != nil {
		panic(err)
	}
	ref := "data.cloudfoundry_stack.s"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(stackDataResource, defaultStacks[0].Name),
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceStackExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", defaultStacks[0].Name),
					),
				},
			},
		})
}

func checkDataSourceStackExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("stack '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]
		description := rs.Primary.Attributes["description"]

		stacks, _, err := session.ClientV2.GetStacks(ccv2.FilterByName(name))
		if err != nil {
			return err
		}
		if len(stacks) == 0 {
			return NotFound
		}
		if err := assertSame(id, stacks[0].GUID); err != nil {
			return err
		}
		if err := assertSame(description, stacks[0].Description); err != nil {
			return err
		}

		return nil
	}
}
