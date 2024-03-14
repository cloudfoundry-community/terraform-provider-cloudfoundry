package cloudfoundry

import (
	"fmt"
	"testing"

	guuid "github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const serviceInstanceSharing = `
data "cloudfoundry_service" "test-service" {
  name = "%s"
}

data "cloudfoundry_user" "u"{
	name = "%s"
	org_id = "%s"
}

resource "cloudfoundry_space" "test-space-1" {
	name = "space-1-%s"
	org = "%s"
	managers = [ data.cloudfoundry_user.u.id ]
	developers = [ data.cloudfoundry_user.u.id ]
	auditors = [ data.cloudfoundry_user.u.id ]
}

resource "cloudfoundry_space" "test-space-2" {
	name = "space-2-%s"
	org = "%s"
	managers = [ data.cloudfoundry_user.u.id ]
	developers = [ data.cloudfoundry_user.u.id ]
	auditors = [ data.cloudfoundry_user.u.id ]
}

resource "cloudfoundry_service_instance" "test-service-instance" {
	name = "test-service-instance-sharing-%s"
	space = resource.cloudfoundry_space.test-space-1.id
	service_plan = data.cloudfoundry_service.test-service.service_plans["%s"]
}

resource "cloudfoundry_service_instance_sharing" "test-service-instance-sharing" {
	service_instance_id = resource.cloudfoundry_service_instance.test-service-instance.id
	space_id = resource.cloudfoundry_space.test-space-2.id
}
`

func TestAccResServiceInstanceSharing_normal(t *testing.T) {
	t.Parallel()
	orgId, _ := defaultTestOrg(t)

	serviceName, _, servicePlan := getTestServiceBrokers(t)
	userName := testSession().Config.User

	testId := guuid.New().String()

	ref := "cloudfoundry_service_instance.test-service-instance"

	resource.Test(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckServiceInstanceDestroyed([]string{"test-service-instance-sharing-%s"}, ref),
				testAccCheckSpaceDestroyed(fmt.Sprintf("space-1-%s", testId)),
				testAccCheckSpaceDestroyed(fmt.Sprintf("space-2-%s", testId)),
			),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(serviceInstanceSharing, serviceName, userName, orgId, testId, orgId, testId, orgId, testId, servicePlan),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceInstanceExists(ref),
						func(s *terraform.State) error {
							for _, rs := range s.RootModule().Resources {
								if rs.Type == "cloudfoundry_service_instance_sharing" {
									return nil
								}
							}
							return fmt.Errorf("resource 'cloudfoundry_service_instance_sharing' not found in terraform state")
						},
					),
				},
			},
		})
}
