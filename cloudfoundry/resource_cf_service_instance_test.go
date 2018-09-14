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

data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}
data "cloudfoundry_service" "mysql" {
    name = "p-mysql"
}

resource "cloudfoundry_service_instance" "mysql" {
	name = "mysql"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.mysql.service_plans["1gb"]}"
	tags = [ "tag-1" , "tag-2" ]
}
`

const serviceInstanceResourceUpdate = `

data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}
data "cloudfoundry_service" "mysql" {
    name = "p-mysql"
}

resource "cloudfoundry_service_instance" "mysql" {
	name = "mysql-updated"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.mysql.service_plans["512mb"]}"
	tags = [ "tag-2", "tag-3", "tag-4" ]
}
`

const serviceInstanceResourceAsyncCreate = `

data "cloudfoundry_domain" "fake-service-broker-domain" {
    name = "%s"
}

data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}
data "cloudfoundry_service" "fake-service" {
	name = "fake-service"
	depends_on = ["cloudfoundry_service_broker.fake-service-broker"]
}

resource "cloudfoundry_route" "fake-service-broker-route" {
	domain = "${data.cloudfoundry_domain.fake-service-broker-domain.id}"
    space = "${data.cloudfoundry_space.space.id}"
	hostname = "fake-service-broker"
	depends_on = ["data.cloudfoundry_domain.fake-service-broker-domain"]
}

resource "cloudfoundry_app" "fake-service-broker" {
    name = "fake-service-broker"
	url = "file://../tests/cf-acceptance-tests/assets/service_broker/"
	space = "${data.cloudfoundry_space.space.id}"
	timeout = 700

	route {
		default_route = "${cloudfoundry_route.fake-service-broker-route.id}"
	}

	depends_on = ["cloudfoundry_route.fake-service-broker-route"]
}

resource "cloudfoundry_service_broker" "fake-service-broker" {
	name = "fake-service-broker"
	url = "http://fake-service-broker.%s"
	username = "admin"
	password = "admin"
	space = "${data.cloudfoundry_space.space.id}"
	depends_on = ["cloudfoundry_app.fake-service-broker"]
}

resource "cloudfoundry_service_instance" "fake-service-instance-with-fake-plan" {
	name = "fake-service-instance-with-fake-plan"
    space = "${data.cloudfoundry_space.space.id}"
	service_plan = "${cloudfoundry_service_broker.fake-service-broker.service_plans["fake-service/fake-plan"]}"
	depends_on = ["cloudfoundry_app.fake-service-broker"]
}

resource "cloudfoundry_service_instance" "fake-service-instance-with-fake-async-plan" {
	name = "fake-service-instance-with-fake-async-plan"
    space = "${data.cloudfoundry_space.space.id}"
	service_plan = "${cloudfoundry_service_broker.fake-service-broker.service_plans["fake-service/fake-async-plan"]}"
	depends_on = ["cloudfoundry_app.fake-service-broker"]
}

resource "cloudfoundry_service_instance" "fake-service-instance-with-fake-async-only-plan" {
	name = "fake-service-instance-with-fake-async-only-plan"
    space = "${data.cloudfoundry_space.space.id}"
	service_plan = "${cloudfoundry_service_broker.fake-service-broker.service_plans["fake-service/fake-async-only-plan"]}"
	depends_on = ["cloudfoundry_app.fake-service-broker"]
}
`

func TestAccServiceInstance_normal(t *testing.T) {

	ref := "cloudfoundry_service_instance.mysql"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceInstanceDestroyed([]string{"mysql", "mysql-updated"}, "data.cloudfoundry_space.space"),
			Steps: []resource.TestStep{

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

func TestAccServiceInstances_withFakePlans(t *testing.T) {

	refFakePlan := "cloudfoundry_service_instance.fake-service-instance-with-fake-plan"
	refFakeAsyncPlan := "cloudfoundry_service_instance.fake-service-instance-with-fake-async-plan"
	refFakeAsyncOnlyPlan := "cloudfoundry_service_instance.fake-service-instance-with-fake-async-only-plan"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceInstanceDestroyed([]string{"fake-service-instance-with-fake-plan", "fake-service-instance-with-fake-async-plan", "fake-service-instance-with-fake-async-only-plan"}, "data.cloudfoundry_space.space"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(serviceInstanceResourceAsyncCreate, defaultAppDomain(), defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						// test fake-plan
						testAccCheckServiceInstanceExists(refFakePlan),
						resource.TestCheckResourceAttr(refFakePlan, "name", "fake-service-instance-with-fake-plan"),
						// test fake-async-plan
						testAccCheckServiceInstanceExists(refFakeAsyncPlan),
						resource.TestCheckResourceAttr(refFakeAsyncPlan, "name", "fake-service-instance-with-fake-async-plan"),
						// test fake-async-only-plan
						testAccCheckServiceInstanceExists(refFakeAsyncOnlyPlan),
						resource.TestCheckResourceAttr(refFakeAsyncOnlyPlan, "name", "fake-service-instance-with-fake-async-only-plan"),
					),
					// ExpectNonEmptyPlan to avoid the following bug in the test
					// --- FAIL: TestAccServiceBroker_async (174.55s)
					//testing.go:513: Step 0 error: After applying this step and refreshing, the plan was not empty:
					//  DIFF:
					//  CREATE: data.cloudfoundry_service.fake-service
					//    name:            "" => "fake-service"
					//    service_plans.%: "" => "<computed>"
					ExpectNonEmptyPlan: true,
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
			return err
		}
		session.Log.DebugMessage(
			"retrieved service instance for resource '%s' with id '%s': %# v",
			resource, id, serviceInstance)

		return nil
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
					continue
				}
			}
			return fmt.Errorf("service instance with name '%s' still exists in cloud foundry", n)
		}
		return nil
	}
}
