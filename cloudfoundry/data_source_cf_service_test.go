package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const serviceDataResource = `

data "cloudfoundry_service" "test" {
    name = "%s"
}
`

func TestAccDataSourceService_normal(t *testing.T) {

	serviceName1, _, servicePlan := getTestServiceBrokers(t)

	ref := "data.cloudfoundry_service.test"

	resource.ParallelTest(t,
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

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]

		services, _, err := session.ClientV2.GetServices(ccv2.FilterEqual(constant.LabelFilter, name))
		if err != nil {
			return err
		}
		if len(services) == 0 {
			return NotFound
		}
		err = assertSame(id, services[0].GUID)

		return err
	}
}
