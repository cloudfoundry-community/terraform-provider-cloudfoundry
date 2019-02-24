package cloudfoundry

import (
	"testing"

	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

func TestAccServiceInstance_importBasic(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)
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
						orgName, spaceName, serviceName1, servicePlan,
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
		session := testAccProvider.Meta().(*cfapi.Session)
		rs, ok := s.RootModule().Resources[serviceInstanceResource]
		if !ok {
			return fmt.Errorf("Service instance '%s' not found in terraform state", spaceResource)
		}
		spaceID := rs.Primary.Attributes["space"]

		for _, n := range names {
			session.Log.DebugMessage("checking ServiceInstance is Destroyed %s", n)
			_, err := session.ServiceManager().FindServiceInstance(n, spaceID)
			if err != nil {
				switch err.(type) {
				case *errors.ModelNotFoundError:
					continue
				default:
					continue
				}
			}
			return fmt.Errorf("service instance with name '%s' still exists in cloud foundry", n)
		}
		return nil
	}
}
