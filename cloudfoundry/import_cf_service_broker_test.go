package cloudfoundry

import (
	"testing"

	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccServiceBroker_importBasic(t *testing.T) {

	serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword, _ := getTestBrokerCredentials(t)

	// Ensure any test artifacts from a
	// failed run are deleted if the exist
	deleteServiceBroker("test")
	deleteServiceBroker("test-renamed")

	resourceName := "cloudfoundry_service_broker.test"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceBrokerDestroyed("test"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(sbResource,
						serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword),
				},

				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"password", "catalog_change", "catalog_hash", "fail_when_catalog_not_accessible"},
				},
			},
		})
}
