package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const asgDataResource = `

data "cloudfoundry_asg" "public" {
    name = "%s"
}
`

func TestAccDataSourceAsg_normal(t *testing.T) {

	defaultAsg := getTestSecurityGroup()
	ref := "data.cloudfoundry_asg.public"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(asgDataResource, defaultAsg),
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceAsgExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", defaultAsg),
					),
				},
			},
		})
}

func checkDataSourceAsgExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("asg '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		asg, _, err := session.ClientV2.GetSecurityGroup(id)
		if err != nil {
			return err
		}
		err = assertEquals(attributes, "name", asg.Name)
		return err
	}
}
