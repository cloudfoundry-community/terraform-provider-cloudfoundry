package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const serviceKeyResource = `

data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}
data "cloudfoundry_service" "rabbitmq" {
    name = "p-rabbitmq"
}

resource "cloudfoundry_service_instance" "rabbitmq" {
	name = "rabbitmq"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.rabbitmq.service_plans["standard"]}"
}

resource "cloudfoundry_service_key" "rabbitmq-key" {
	name = "rabbitmq-key"
	service_instance = "${cloudfoundry_service_instance.rabbitmq.id}"

	params {
		"key1" = "aaaa"
		"key2" = "bbbb"
	}
}
`

func TestAccServiceKey_normal(t *testing.T) {

	ref := "cloudfoundry_service_key.rabbitmq-key"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceKeyDestroyed("rabbitmq-key", "cloudfoundry_service_instance.rabbitmq"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: serviceKeyResource,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceKeyExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "rabbitmq-key"),
					),
				},
			},
		})
}

func testAccCheckServiceKeyExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service instance '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var serviceKey cfapi.CCServiceKey

		sm := session.ServiceManager()
		if serviceKey, err = sm.ReadServiceKey(id); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved service instance for resource '%s' with id '%s': %# v",
			resource, id, serviceKey)

		if err = assertEquals(attributes, "name", serviceKey.Name); err != nil {
			return err
		}

		//normalized := normalizeMap(serviceKey.Credentials, make(map[string]interface{}), "", "_")
		return assertMapEquals("credentials", attributes, serviceKey.Credentials)
	}
}

func testAccCheckServiceKeyDestroyed(name, serviceInstance string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[serviceInstance]
		if !ok {
			return fmt.Errorf("service instance '%s' not found in terraform state", spaceResource)
		}

		session.Log.DebugMessage("checking ServiceKey is Destroyed %s", name)

		if _, err := session.ServiceManager().FindServiceKey(name, rs.Primary.ID); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil

			default:
				return err
			}
		}
		return fmt.Errorf("service instance with name '%s' still exists in cloud foundry", name)
	}
}
