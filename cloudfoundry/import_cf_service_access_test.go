package cloudfoundry

import (
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccServiceAccess_importBasic(t *testing.T) {
	resourceName := "cf_service_access.redis-access"

	user, password := getRedisBrokerCredentials()
	deleteServiceBroker("p-redis")

	var servicePlanAccessGUID string

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceAccessDestroyed(servicePlanAccessGUID),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(saResource,
						defaultSysDomain(), user, password, defaultPcfDevOrgID()),
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
