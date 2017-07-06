package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const servicePlanDataResource = `

data "cf_service" "redis" {
    name = "p-redis"
}
data "cf_service_plan" "redis" {
    name = "shared-vm"
	service = "${data.cf_service.redis.id}"
}
`

func TestAccDataSourceServicePlan_normal(t *testing.T) {

	ref := "data.cf_service_plan.redis"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: servicePlanDataResource,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceServicePlanExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "shared-vm"),
					),
				},
			},
		})
}

func checkDataSourceServicePlanExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service plan '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]
		service := rs.Primary.Attributes["service"]

		var (
			planID string
			err    error
		)

		planID, err = session.ServiceManager().FindServicePlanID(service, name)
		if err != nil {
			return err
		}
		if err := assertSame(id, planID); err != nil {
			return err
		}

		return nil
	}
}
