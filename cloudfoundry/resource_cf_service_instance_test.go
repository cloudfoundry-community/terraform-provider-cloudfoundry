package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const serviceInstanceResourceCreate = `

data "cf_org" "org" {
    name = "pcfdev-org"
}
data "cf_space" "space" {
    name = "pcfdev-space"
	org = "${data.cf_org.org.id}"
}
data "cf_service" "mysql" {
    name = "p-mysql"
}

resource "cf_service_instance" "mysql" {
	name = "mysql"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.mysql.service_plans["1gb"]}"
	tags = [ "tag-1" , "tag-2" ]
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
data "cf_service" "mysql" {
    name = "p-mysql"
}

resource "cf_service_instance" "mysql" {
	name = "mysql-updated"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.mysql.service_plans["512mb"]}"
	tags = [ "tag-2", "tag-3", "tag-4" ]
}
`

const serviceInstanceResourceAsyncCreate = `

data "cf_org" "org" {
    name = "pcfdev-org"
}
data "cf_space" "space" {
    name = "pcfdev-space"
	org = "${data.cf_org.org.id}"
}
data "cf_service" "test-service" {
	name = "test-service"
	
	depends_on = ["cf_service_broker.test-service-broker"]
}

resource "cf_app" "test-service-broker" {
    name = "test-service-broker"
	url = "file://service_broker/"
	space = "${data.cf_space.space.id}"
}

resource "cf_service_broker" "test-service-broker {
	name = "test-service-broker"
	url = "http://test-service-broker.local.pcfdev.io"
	depends_on = ["cf_app.test-service-broker"]
}

resource "cf_service_instance" "test-service-instance" {
	name = "test-service-instance"
    space = "${data.cf_space.space.id}"
	service_plan = "${cf_service_broker.service_plans["test-service/test-async-only-plan"]}"
	depends_on = ["cf_app.test-service-broker"]
}
`

// TODO - Add Service Broker with async. service plans
func TestAccServiceInstance_normal(t *testing.T) {

	ref := "cf_service_instance.mysql"
	refAsync := "cf_service_instance.test-service-instance"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceInstanceDestroyed([]string{"mysql", "mysql-updated", "test-service-instance"}, "data.cf_space.space"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: serviceInstanceResourceAsyncCreate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceInstanceExists(refAsync),
						resource.TestCheckResourceAttr(refAsync, "name", "test-service-instance"),
					),
				},
				resource.TestStep{
					Config: serviceInstanceResourceCreate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceInstanceExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "mysql"),
						resource.TestCheckResourceAttr(
							ref, "tags.#", "2"),
						resource.TestCheckResourceAttr(
							ref, "tags.0", "tag-1"),
						resource.TestCheckResourceAttr(
							ref, "tags.1", "tag-2"),
					),
				},

				resource.TestStep{
					Config: serviceInstanceResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceInstanceExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "mysql-updated"),
						resource.TestCheckResourceAttr(
							ref, "tags.#", "3"),
						resource.TestCheckResourceAttr(
							ref, "tags.0", "tag-2"),
						resource.TestCheckResourceAttr(
							ref, "tags.1", "tag-3"),
						resource.TestCheckResourceAttr(
							ref, "tags.2", "tag-4"),
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

func testAccCheckServiceInstanceDestroyed(names []string, spaceResource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[spaceResource]
		if !ok {
			return fmt.Errorf("space '%s' not found in terraform state", spaceResource)
		}

		for _, n := range names {

			session.Log.DebugMessage("checking ServiceInstance is Destroyed %s", n)

			if _, err := session.ServiceManager().FindServiceInstance(n, rs.Primary.ID); err != nil {
				switch err.(type) {
				case *errors.ModelNotFoundError:
					return nil

				default:
					break
				}
			}
			return fmt.Errorf("service instance with name '%s' still exists in cloud foundry", n)
		}
		return nil
	}
}
