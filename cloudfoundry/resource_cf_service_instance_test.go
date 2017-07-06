package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const serviceInstanceResourceCreate = `

data "cf_org" "org" {
    name = "pcfdev-org"
}
data "cf_space" "space" {
    name = "pcfdev-space"
	org = "${data.cf_org.org.id}"
}

data "cf_service" "redis" {
    name = "p-redis"
}
data "cf_service_plan" "redis" {
    name = "shared-vm"
    service = "${data.cf_service.redis.id}"
}

resource "cf_service_instance" "redis1" {
	name = "redis1"
    space = "${cf_space.space.id}"
    servicePlan = "${data.cf_service_plan.redis.id}"
}
`

const serviceInstanceResourceUpdate = `

data "cf_org" "org" {
    name = "pcfdev-org"
}
data "cf_space" "space" {
    name = "pcfdev-space"
	org = "${data.cf_org.org.id}"
}

data "cf_service" "redis" {
    name = "p-redis"
}
data "cf_service_plan" "redis" {
    name = "shared-vm"
    service = "${data.cf_service.redis.id}"
}

resource "cf_service_instance" "redis1" {
	name = "redis-new-name"
    space = "${cf_space.space.id}"
    servicePlan = "${data.cf_service_plan.redis.id}"
	tags = [ "redis" , "data-grid" ]
}
`

func TestAccServiceInstance_normal(t *testing.T) {

	ref := "cf_service_instance.redis1"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceInstanceDestroyed("redis1", "cf_space.space1"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: serviceInstanceResourceCreate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceInstanceExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "redis1"),
						resource.TestCheckNoResourceAttr(
							ref, "tags"),
					),
				},

				resource.TestStep{
					Config: serviceInstanceResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceInstanceExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "redis-new-name"),
						resource.TestCheckResourceAttr(
							ref, "tags.#", "2"),
						resource.TestCheckResourceAttr(
							ref, "tags.0", "redis"),
						resource.TestCheckResourceAttr(
							ref, "tags.1", "data-grid"),
					),
				},
			},
		})
}

func testAccCheckServiceInstanceExists(resource string) resource.TestCheckFunc {

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

		var (
			serviceInstance cfapi.CCServiceInstance
		)

		sm := session.ServiceManager()
		if serviceInstance, err = sm.ReadServiceInstance(id); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved service instance for resource '%s' with id '%s': %# v",
			resource, id, serviceInstance)

		return
	}
}

func testAccCheckServiceInstanceDestroyed(name string, spaceResource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[spaceResource]
		if !ok {
			return fmt.Errorf("space '%s' not found in terraform state", spaceResource)
		}

		session.Log.DebugMessage("checking ServiceInstance is Destroyed %s", name)

		if _, err := session.ServiceManager().FindServiceInstance(name, rs.Primary.ID); err != nil {
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
