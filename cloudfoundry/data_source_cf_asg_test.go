package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const asgDataResource = `

data "cf_asg" "public" {
    name = "%s"
}
`

func TestAccDataSourceAsg_normal(t *testing.T) {

	defaultAsg := getDefaultSecurityGroup()
	ref := "data.cf_asg.public"

	resource.Test(t,
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

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("asg '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		asg, err := session.ASGManager().GetASG(id)
		if err != nil {
			return err
		}
		err = assertEquals(attributes, "name", asg.Name)
		return err
	}
}
