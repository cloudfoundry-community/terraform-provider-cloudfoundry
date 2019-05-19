package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const sbResource = `

resource "cloudfoundry_service_broker" "test" {
	name = "test"
	url = "%s"
	username = "%s"
	password = "%s"
}
`

const sbResourceUpdate = `

resource "cloudfoundry_service_broker" "test" {
	name = "test-renamed"
	url = "%s"
	username = "%s"
	password = "%s"
}
`

func TestAccServiceBroker_normal(t *testing.T) {

	serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword, serviceBrokerPlanPath := getTestBrokerCredentials(t)

	// Ensure any test artifacts from a
	// failed run are deleted if the exist
	deleteServiceBroker("test")
	deleteServiceBroker("test-renamed")

	ref := "cloudfoundry_service_broker.test"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceBrokerDestroyed("test"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(sbResource,
						serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceBrokerExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "test"),
						resource.TestCheckResourceAttr(
							ref, "url", serviceBrokerURL),
						resource.TestCheckResourceAttr(
							ref, "username", serviceBrokerUser),
						resource.TestCheckResourceAttrSet(
							ref, "service_plans."+serviceBrokerPlanPath),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(sbResourceUpdate,
						serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceBrokerExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "test-renamed"),
					),
				},
			},
		})
}

func testAccCheckServiceBrokerExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service broker '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		serviceBroker, _, err := session.ClientV2.GetServiceBroker(id)
		if err != nil {
			return err
		}

		if err := assertEquals(attributes, "name", serviceBroker.Name); err != nil {
			return err
		}
		if err := assertEquals(attributes, "url", serviceBroker.BrokerURL); err != nil {
			return err
		}
		if err := assertEquals(attributes, "username", serviceBroker.AuthUsername); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckServiceBrokerDestroyed(name string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)
		sbs, _, err := session.ClientV2.GetServiceBrokers(ccv2.FilterByName(name))
		if err != nil {
			return err
		}
		if len(sbs) > 0 {
			return fmt.Errorf("service broker with name '%s' still exists in cloud foundry", name)
		}
		return nil
	}
}
