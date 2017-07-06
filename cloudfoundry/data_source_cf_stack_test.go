package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const stackDataResource = `

data "cf_stack" "s" {
    name = "cflinuxfs2"
}
`

func TestAccDataSourceStack_normal(t *testing.T) {

	ref := "data.cf_stack.s"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: stackDataResource,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceStackExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "cflinuxfs2"),
					),
				},
			},
		})
}

func checkDataSourceStackExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("stack '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]
		description := rs.Primary.Attributes["description"]

		var (
			err   error
			stack cfapi.CCStack
		)

		stack, err = session.StackManager().FindStackByName(name)
		if err != nil {
			return err
		}
		if err := assertSame(id, stack.ID); err != nil {
			return err
		}
		if err := assertSame(description, stack.Description); err != nil {
			return err
		}

		return nil
	}
}
