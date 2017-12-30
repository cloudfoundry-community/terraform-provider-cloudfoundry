package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const saResource = `

resource "cf_service_broker" "redis" {
	name = "test-redis"
	url = "https://redis-broker.%s"
	username = "%s"
	password = "%s"
}

resource "cf_service_access" "redis-access" {
	plan = "${cf_service_broker.redis.service_plans["p-redis/shared-vm"]}"
	org = "%s"
}
`

func TestAccServiceAccess_normal(t *testing.T) {

	user, password := getRedisBrokerCredentials()
	deleteServiceBroker("p-redis")

	var servicePlanAccessGUID string
	ref := "cf_service_access.redis-access"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceAccessDestroyed(servicePlanAccessGUID),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(saResource,
						defaultSysDomain(), user, password, defaultPcfDevOrgID()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceAccessExists(ref,
							func(guid string) {
								servicePlanAccessGUID = guid
							}),
						resource.TestCheckResourceAttrSet(
							ref, "plan"),
						resource.TestCheckResourceAttr(
							ref, "org", defaultPcfDevOrgID()),
					),
				},
			},
		})
}

func testAccCheckServiceAccessExists(resource string,
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

func testAccCheckServiceAccessDestroyed(servicePlanAccessGUID string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		_, _, err := session.ServiceManager().ReadServicePlanAccess(servicePlanAccessGUID)
		if err == nil {
			return fmt.Errorf("service plan access with guid '%s' still exists in cloud foundry", servicePlanAccessGUID)
		}
		return nil
	}
}
