package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/models"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const serviceDataResource = `

data "cf_service" "redis" {
    name = "p-redis"
}
`

func TestAccDataSourceService_normal(t *testing.T) {

	ref := "data.cf_service.redis"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: serviceDataResource,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceServiceExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "p-redis"),
						resource.TestCheckResourceAttrSet(
							ref, "service_plans.shared-vm"),
					),
				},
			},
		})
}

func checkDataSourceServiceExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]

		var (
			err     error
			service models.ServiceOffering
		)

		service, err = session.ServiceManager().FindServiceByName(name)
		if err != nil {
			return err
		}
		if err := assertSame(id, service.GUID); err != nil {
			return err
		}

		return nil
	}
}
