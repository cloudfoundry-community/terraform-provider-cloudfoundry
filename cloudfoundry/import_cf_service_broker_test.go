package cloudfoundry

import (
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
)

const sbImportedResource = `

resource "cf_service_broker" "redis" {
	name = "test-redis"
	url = "https://redis-broker.%s"
	username = "%s"
	password = "%s"
}
`

func TestAccServiceBroker_importBasic(t *testing.T) {
	resourceName := "cf_service_broker.redis"

	user, password := getRedisBrokerCredentials()
	deleteServiceBroker("p-redis")

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceBrokerDestroyed("test-redis"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(sbImportedResource,
						defaultSysDomain(), user, password),
				},

				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"password"},
				},
			},
		})
}
