package cloudfoundry

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

const saResource = `
resource "cloudfoundry_service_broker" "test" {
	name = "test"
	url = "%s"
	username = "%s"
	password = "%s"
}

resource "cloudfoundry_service_plan_access" "test-access" {
	plan = "${cloudfoundry_service_broker.test.service_plans["%s"]}"
	org = "%s"
}
`

const saResourceUpdateTrue = `
resource "cloudfoundry_service_broker" "test" {
	name = "test"
	url = "%s"
	username = "%s"
	password = "%s"
}

resource "cloudfoundry_service_plan_access" "test-access" {
	plan = "${cloudfoundry_service_broker.test.service_plans["%s"]}"
	public = true
}
`

const saResourceUpdateFalse = `
resource "cloudfoundry_service_broker" "test" {
	name = "test"
	url = "%s"
	username = "%s"
	password = "%s"
}

resource "cloudfoundry_service_plan_access" "test-access" {
	plan = "${cloudfoundry_service_broker.test.service_plans["%s"]}"
	public = false
}
`

const saResourceError = `
resource "cloudfoundry_service_broker" "test" {
	name = "test"
	url = "%s"
	username = "%s"
	password = "%s"
}

resource "cloudfoundry_service_plan_access" "test-access" {
	plan = "${cloudfoundry_service_broker.test.service_plans["%s"]}"
	org = "%s"
	public = true
}
`

func TestAccServicePlanAccess_normal(t *testing.T) {

	serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword, serviceBrokerPlanPath := getTestBrokerCredentials(t)

	// Ensure any test artifacts from a
	// failed run are deleted if the exist
	deleteServiceBroker("test")

	orgID, _ := defaultTestOrg(t)
	ref := "cloudfoundry_service_plan_access.test-access"

	var servicePlanAccessGUID string

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServicePlanAccessDestroyed(servicePlanAccessGUID),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(saResource,
						serviceBrokerURL,
						serviceBrokerUser,
						serviceBrokerPassword,
						serviceBrokerPlanPath,
						orgID),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServicePlanAccessExists(ref,
							func(guid string) {
								servicePlanAccessGUID = guid
							}),
						resource.TestCheckResourceAttrSet(ref, "plan"),
						resource.TestCheckResourceAttr(ref, "org", orgID),
					),
				},
				resource.TestStep{
					Config: fmt.Sprintf(saResourceUpdateTrue,
						serviceBrokerURL,
						serviceBrokerUser,
						serviceBrokerPassword,
						serviceBrokerPlanPath),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServicePlan(ref),
						resource.TestCheckResourceAttrSet(ref, "plan"),
						resource.TestCheckResourceAttr(ref, "public", "true"),
					),
				},
				resource.TestStep{
					Config: fmt.Sprintf(saResourceUpdateFalse,
						serviceBrokerURL,
						serviceBrokerUser,
						serviceBrokerPassword,
						serviceBrokerPlanPath),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServicePlan(ref),
						resource.TestCheckResourceAttrSet(ref, "plan"),
						resource.TestCheckResourceAttr(ref, "public", "false"),
					),
				},
			},
		})
}

func TestAccServicePlanAccess_error(t *testing.T) {

	serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword, serviceBrokerPlanPath := getTestBrokerCredentials(t)

	// Ensure any test artifacts from a
	// failed run are deleted if the exist
	deleteServiceBroker("test")

	orgID, _ := defaultTestOrg(t)
	var servicePlanAccessGUID string

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServicePlanAccessDestroyed(servicePlanAccessGUID),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(saResourceError,
						serviceBrokerURL,
						serviceBrokerUser,
						serviceBrokerPassword,
						serviceBrokerPlanPath,
						orgID),
					ExpectError: regexp.MustCompile("\"org\": conflicts with public"),
				},
			},
		})
}

func testAccCheckServicePlanAccessExists(resource string,
	setServicePlanAccessGUID func(string)) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)
		sm := session.ServiceManager()

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service access resource '%s' not found in terraform state", rs)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		setServicePlanAccessGUID(id)

		plan, org, err := sm.ReadServicePlanAccess(id)
		if err != nil {
			return err
		}
		if err := assertEquals(attributes, "plan", plan); err != nil {
			return err
		}
		if err := assertEquals(attributes, "org", org); err != nil {
			return err
		}

		return
	}
}

func testAccCheckServicePlan(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		session := testAccProvider.Meta().(*cfapi.Session)
		sm := session.ServiceManager()
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service access resource '%s' not found in terraform state", rs)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		plan, err := sm.ReadServicePlan(id)
		if err != nil {
			return err
		}
		if err := assertEquals(attributes, "plan", id); err != nil {
			return err
		}
		if err := assertEquals(attributes, "public", plan.Public); err != nil {
			return err
		}
		return
	}
}

func testAccCheckServicePlanAccessDestroyed(servicePlanAccessGUID string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		_, _, err := session.ServiceManager().ReadServicePlanAccess(servicePlanAccessGUID)
		if err == nil {
			return fmt.Errorf("service plan access with guid '%s' still exists in cloud foundry", servicePlanAccessGUID)
		}
		return nil
	}
}
