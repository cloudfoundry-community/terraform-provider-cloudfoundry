package cloudfoundry

import (
	"fmt"
	"log"
	"testing"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const serviceInstanceResourceCreate = `
data "cloudfoundry_service" "test-service" {
  name = "%s"
}

resource "cloudfoundry_service_instance" "test-service-instance" {
  name = "test-service-instance"
  space = "%s"
  service_plan = "${data.cloudfoundry_service.test-service.service_plans["%s"]}"
  tags = [ "tag-1" , "tag-2" ]
  labels = {
	instance-name = "test-service-instance"
  }
}
`

const serviceInstanceResourceUpdate = `
data "cloudfoundry_service" "test-service" {
  name = "%s"
}

resource "cloudfoundry_service_instance" "test-service-instance" {
  name = "test-service-instance-updated"
  space = "%s"
  service_plan = "${data.cloudfoundry_service.test-service.service_plans["%s"]}"
  tags = [ "tag-2", "tag-3", "tag-4" ]
  labels = {
	instance-name = "%s"
  }
}
`

const serviceInstanceResourceAsyncCreate = `
data "cloudfoundry_service" "fake-service" {
  name = "${keys(cloudfoundry_service_broker.fake-service-broker.services)[0]}"
}

resource "cloudfoundry_service_instance" "fake-service-instance-with-fake-plan" {
  name = "fake-service-instance-with-fake-plan"
  space = "%s"
  service_plan = "${data.cloudfoundry_service.fake-service.service_plans["fake-plan"]}"
  depends_on = ["cloudfoundry_service_broker.fake-service-broker"]
}

resource "cloudfoundry_service_instance" "fake-service-instance-with-fake-async-plan" {
  name = "fake-service-instance-with-fake-async-plan"
  space = "%s"
  service_plan = "${data.cloudfoundry_service.fake-service.service_plans["fake-async-plan"]}"
  depends_on = ["cloudfoundry_service_broker.fake-service-broker"]
}

%s
`

const fakeServiceBroker = `

data "cloudfoundry_domain" "fake-service-broker-domain" {
  name = "%s"
}

resource "cloudfoundry_route" "fake-service-broker-route" {
  domain = "${data.cloudfoundry_domain.fake-service-broker-domain.id}"
  space = "%s"
  hostname = "fake-service-broker"
}

resource "cloudfoundry_app" "fake-service-broker" {
  name = "fake-service-broker"
  path = "file://../tests/cf-acceptance-tests/assets/cats-service-broker.zip"
  space = "%s"
  timeout = 700

  routes {
    route = "${cloudfoundry_route.fake-service-broker-route.id}"
  }
}

resource "cloudfoundry_service_broker" "fake-service-broker" {
  name = "fake-service-broker"
  url = "https://fake-service-broker.%s"
  username = "admin"
  password = "admin"
  space = "%s"
  depends_on = ["cloudfoundry_app.fake-service-broker"]
}
`

func TestAccResServiceInstance_normal(t *testing.T) {

	spaceId, _ := defaultTestSpace(t)
	serviceName1, _, servicePlan := getTestServiceBrokers(t)
	labelVal := "new-service-label-updated"

	ref := "cloudfoundry_service_instance.test-service-instance"

	resource.Test(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			CheckDestroy: testAccCheckServiceInstanceDestroyed(
				[]string{
					"test-service",
					"test-service-updated"},
				ref),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(serviceInstanceResourceCreate,
						serviceName1, spaceId, servicePlan,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceInstanceExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "test-service-instance"),
						resource.TestCheckResourceAttr(
							ref, "tags.#", "2"),
						resource.TestCheckResourceAttr(
							ref, "tags.0", "tag-1"),
						resource.TestCheckResourceAttr(
							ref, "tags.1", "tag-2"),
						resource.TestCheckResourceAttr(
							ref, "labels.instance-name", "test-service-instance"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(serviceInstanceResourceUpdate,
						serviceName1, spaceId, servicePlan, labelVal,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceInstanceExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "test-service-instance-updated"),
						resource.TestCheckResourceAttr(
							ref, "tags.#", "3"),
						resource.TestCheckResourceAttr(
							ref, "tags.0", "tag-2"),
						resource.TestCheckResourceAttr(
							ref, "tags.1", "tag-3"),
						resource.TestCheckResourceAttr(
							ref, "tags.2", "tag-4"),
						resource.TestCheckResourceAttr(
							ref, "labels.instance-name", labelVal),
					),
				},
				resource.TestStep{
					Config: fmt.Sprintf(serviceInstanceResourceCreate,
						serviceName1, spaceId, servicePlan,
					),
					Check: testAccCheckServiceInstanceMetadataExists(ref),
				},
			},
		})
}

func TestAccResServiceInstances_withFakePlans(t *testing.T) {

	spaceId, _ := defaultTestSpace(t)
	appDomain := defaultAppDomain()

	refFakePlan := "cloudfoundry_service_instance.fake-service-instance-with-fake-plan"
	refFakeAsyncPlan := "cloudfoundry_service_instance.fake-service-instance-with-fake-async-plan"
	resource.Test(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			CheckDestroy: testAccCheckServiceInstanceDestroyed(
				[]string{
					"fake-service-instance-with-fake-plan",
					"fake-service-instance-with-fake-async-plan"},
				refFakePlan),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(serviceInstanceResourceAsyncCreate,
						spaceId, spaceId,
						fmt.Sprintf(fakeServiceBroker, appDomain, spaceId, spaceId, appDomain, spaceId),
					),
					Check: resource.ComposeTestCheckFunc(
						// test fake-plan
						testAccCheckServiceInstanceExists(refFakePlan),
						resource.TestCheckResourceAttr(refFakePlan, "name", "fake-service-instance-with-fake-plan"),
						// test fake-async-plan
						testAccCheckServiceInstanceExists(refFakeAsyncPlan),
						resource.TestCheckResourceAttr(refFakeAsyncPlan, "name", "fake-service-instance-with-fake-async-plan"),
					),
				},
			},
		})
}

func testAccCheckServiceInstanceMetadataExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		// Retrieve the currently active session object
		session := testAccProvider.Meta().(*managers.Session)

		// rs : represents the resource from Terraform statefile
		rs, ok := s.RootModule().Resources[resource]

		if !ok {
			return fmt.Errorf("service instance '%s' not found in terraform state", resource)
		}

		// It retrieves the ID of the resource from the Terraform state.
		GUID := rs.Primary.ID

		serviceInstances, _, _, err := session.ClientV3.GetServiceInstances(ccv3.Query{
			Key:    ccv3.GUIDFilter,
			Values: []string{GUID},
		})
		if err != nil {
			return err
		}

		if len(serviceInstances) == 0 {
			return fmt.Errorf("Service instance with guid: %s not found", GUID)
		} else {
			// _ is used to ignore the i of this loop
			for _, instance := range serviceInstances {
				metadata := instance.Metadata
				labelVal, found := metadata.Labels["instance-name"]
				log.Printf("!!!! Found service instance : %+v", serviceInstances)
				if found {
					fmt.Printf("Label 'Instance-name' is: %v\n", labelVal)
					return nil
				} else {
					return fmt.Errorf("Label 'instance-name' does not exist in this resource\n")
				}
			}

		}

		return fmt.Errorf("Unexpected condition, no return hit")
	}
}

func testAccCheckServiceInstanceExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service instance '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		_, _, err := session.ClientV2.GetServiceInstance(id)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckServiceInstanceDestroyed(names []string, testResource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
		rs, ok := s.RootModule().Resources[testResource]
		if !ok {
			return fmt.Errorf("the service instance '%s' not found in terraform state", testResource)
		}

		for _, n := range names {
			sis, _, err := session.ClientV2.GetServiceInstances(ccv2.FilterByName(n), ccv2.FilterEqual(constant.SpaceGUIDFilter, rs.Primary.Attributes["space"]))
			if err != nil {
				return err
			}
			if len(sis) > 0 {
				return fmt.Errorf("service instance with name '%s' still exists in cloud foundry", n)
			}
		}
		return nil
	}
}
