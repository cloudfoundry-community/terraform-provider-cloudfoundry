package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const sbResource = `

resource "cf_service_broker" "redis" {
	name = "test-redis"
	url = "https://redis-broker.%s"
	username = "%s"
	password = "%s"
  visibilities = [
   {
     service = "p-redis"
     private = [ "shared-vm" ]
   }
  ]
}
`

const sbResourceUpdate = `

resource "cf_service_broker" "redis" {
	name = "test-redis-renamed"
	url = "https://redis-broker.%s"
	username = "%s"
	password = "%s"
  visibilities = [
   {
     service = "p-redis"
     public  = [ "shared-vm" ]
   }
  ]
}
`

func TestAccServiceBroker_normal(t *testing.T) {

	user, password := getRedisBrokerCredentials()
	deleteServiceBroker("p-redis")

	ref := "cf_service_broker.redis"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceBrokerDestroyed("test-redis"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(sbResource,
						defaultSysDomain(), user, password),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceBrokerExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "test-redis"),
						resource.TestCheckResourceAttr(
							ref, "url", "https://redis-broker."+defaultSysDomain()),
						resource.TestCheckResourceAttr(
							ref, "username", "admin"),
						resource.TestCheckResourceAttr(
							ref, "visibilities.#", "1"),
						resource.TestCheckResourceAttr(
							ref, fmt.Sprintf("visibilities.%d.service",
								hashVisibilityObj("p-redis", nil, []string{"shared-vm"})),
							"p-redis"),
						resource.TestCheckResourceAttr(
							ref, fmt.Sprintf("visibilities.%d.public.#",
								hashVisibilityObj("p-redis", nil, []string{"shared-vm"})),
							"0"),
						resource.TestCheckResourceAttr(
							ref, fmt.Sprintf("visibilities.%d.private.#",
								hashVisibilityObj("p-redis", nil, []string{"shared-vm"})),
							"1"),
						resource.TestCheckResourceAttr(
							ref, fmt.Sprintf("visibilities.%d.private.%d",
								hashVisibilityObj("p-redis", nil, []string{"shared-vm"}),
								hashcode.String("shared-vm")),
							"shared-vm"),
						resource.TestCheckResourceAttrSet(
							ref, "service_plans.p-redis/shared-vm"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(sbResourceUpdate,
						defaultSysDomain(), user, password),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceBrokerExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "test-redis-renamed"),
						resource.TestCheckResourceAttr(
							ref, "visibilities.#", "1"),
						resource.TestCheckResourceAttr(
							ref, fmt.Sprintf("visibilities.%d.service",
								hashVisibilityObj("p-redis", []string{"shared-vm"}, nil)),
							"p-redis"),
						resource.TestCheckResourceAttr(
							ref, fmt.Sprintf("visibilities.%d.private.#",
								hashVisibilityObj("p-redis", []string{"shared-vm"}, nil)),
							"0"),
						resource.TestCheckResourceAttr(
							ref, fmt.Sprintf("visibilities.%d.public.#",
								hashVisibilityObj("p-redis", []string{"shared-vm"}, nil)),
							"1"),
						resource.TestCheckResourceAttr(
							ref, fmt.Sprintf("visibilities.%d.public.%d",
								hashVisibilityObj("p-redis", []string{"shared-vm"}, nil),
								hashcode.String("shared-vm")),
							"shared-vm"),
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
