package cloudfoundry

import (
	"testing"

	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccServicePlanAccess_importBasic(t *testing.T) {

	serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword, serviceBrokerPlanPath := getTestBrokerCredentials(t)

	// Ensure any test artifacts from a
	// failed run are deleted if the exist
	deleteServiceBroker("test")

	orgID, _ := defaultTestOrg(t)
	resourceName := "cloudfoundry_service_plan_access.test-access"

	var servicePlanAccessGUID string

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServicePlanAccessDestroyed(servicePlanAccessGUID),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(saResource,
						serviceBrokerURL,
						serviceBrokerUser,
						serviceBrokerPassword,
						serviceBrokerPlanPath,
						orgID),
				},
				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
