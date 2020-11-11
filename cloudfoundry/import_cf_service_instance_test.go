package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccServiceInstance_importBasic(t *testing.T) {

	spaceId, _ := defaultTestSpace(t)
	serviceName1, _, servicePlan := getTestServiceBrokers(t)

	resourceName := "cloudfoundry_service_instance.test-service-instance"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: testAccCheckServiceInstanceDestroyedImportState(
				[]string{"test-service"},
				resourceName),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(serviceInstanceResourceCreate,
						serviceName1, spaceId, servicePlan,
					),
				},
				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"recursive_delete"},
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceInstanceExists(resourceName),
						resource.TestCheckResourceAttr(
							resourceName, "name", "test-service-instance"),
						resource.TestCheckResourceAttr(
							resourceName, "tags.#", "2"),
						resource.TestCheckResourceAttr(
							resourceName, "tags.0", "tag-1"),
						resource.TestCheckResourceAttr(
							resourceName, "tags.1", "tag-2"),
						resource.TestCheckResourceAttr(
							resourceName, "json_params", ""),
					),
				},
			},
		})
}

// after checking import state doesn't have data resource space, only the imported service instance.
// check must use id imported instead of using one found in first state (before importing)
func testAccCheckServiceInstanceDestroyedImportState(names []string, serviceInstanceResource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
		rs, ok := s.RootModule().Resources[serviceInstanceResource]
		if !ok {
			return fmt.Errorf("Service instance '%s' not found in terraform state", spaceResource)
		}
		spaceID := rs.Primary.Attributes["space"]

		for _, n := range names {
			sis, _, err := session.ClientV2.GetServiceInstances(ccv2.FilterByName(n), ccv2.FilterEqual(constant.SpaceGUIDFilter, spaceID))
			if err != nil {
				return err
			}
			if len(sis) > 0 {
				return fmt.Errorf("service instance with name '%s' still exists in cloud foundry", n)
			}
		}
		return nil
	}
}
