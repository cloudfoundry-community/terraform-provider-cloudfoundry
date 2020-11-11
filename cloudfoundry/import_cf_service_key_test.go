package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccServiceKey_importBasic(t *testing.T) {
	spaceId, _ := defaultTestSpace(t)
	serviceName1, _, servicePlan := getTestServiceBrokers(t)

	resourceName := "cloudfoundry_service_key.test-service-instance-key"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: testAccCheckServiceKeyDestroyed(
				"test-service-instance-key",
				"cloudfoundry_service_instance.test-service-instance"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(serviceKeyResource,
						serviceName1, spaceId, servicePlan),
				},

				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"params"},
				},
			},
		})
}
