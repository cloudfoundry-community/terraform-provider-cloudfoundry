package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const sbResource = `

resource "cf_service_broker" "redis" {
	name = "test-redis"
	url = "http://redis-broker.%s"
	username = "admin"
	password = "admin"
}
`

const sbResourceUpdate = `

resource "cf_service_broker" "redis" {
	name = "test-redis-renamed"
	url = "http://redis-broker.%s"
	username = "admin"
	password = "admin"
}
`

func TestAccServiceBroker_normal(t *testing.T) {

	deleteServiceBroker("p-redis")

	ref := "cf_service_broker.redis"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceBrokerDestroyed("test-redis"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(sbResource, defaultDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceBrokerExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "test-redis"),
						resource.TestCheckResourceAttr(
							ref, "url", "http://redis-broker."+defaultDomain()),
						resource.TestCheckResourceAttr(
							ref, "username", "admin"),
						resource.TestCheckResourceAttrSet(
							ref, "service_plans.p-redis/shared-vm"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(sbResourceUpdate, defaultDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceBrokerExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "test-redis-renamed"),
					),
				},
			},
		})
}

func testAccCheckServiceBrokerExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service broker '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var (
			serviceBroker cfapi.CCServiceBroker
		)

		sm := session.ServiceManager()
		if serviceBroker, err = sm.ReadServiceBroker(id); err != nil {
			return
		}

		if err := assertEquals(attributes, "name", serviceBroker.Name); err != nil {
			return err
		}
		if err := assertEquals(attributes, "url", serviceBroker.BrokerURL); err != nil {
			return err
		}
		if err := assertEquals(attributes, "username", serviceBroker.AuthUserName); err != nil {
			return err
		}

		return
	}
}

func testAccCheckServiceBrokerDestroyed(name string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)
		if _, err := session.ServiceManager().GetServiceBrokerID(name); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}

		return fmt.Errorf("service broker with name '%s' still exists in cloud foundry", name)
	}
}
