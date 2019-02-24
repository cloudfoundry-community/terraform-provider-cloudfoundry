package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/models"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

const serviceDataResource = `

data "cloudfoundry_service" "test" {
    name = "%s"
}
`

func TestAccDataSourceService_normal(t *testing.T) {

	serviceName1, _, servicePlan := getTestServiceBrokers(t)

	ref := "data.cloudfoundry_service.test"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(serviceDataResource,
						serviceName1),
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceServiceExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", serviceName1),
						resource.TestCheckResourceAttrSet(
							ref, "service_plans."+servicePlan),
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
		err = assertSame(id, service.GUID)

		return err
	}
}
