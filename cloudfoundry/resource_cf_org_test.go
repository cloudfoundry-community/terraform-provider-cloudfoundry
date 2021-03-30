package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const orgResource = `

resource "cloudfoundry_org_quota" "runaway" {
	name = "runaway_test"
    allow_paid_service_plans = true
    instance_memory = -1
    total_app_instances = -1
    total_memory = 204800
    total_routes = 2000
    total_services = -1
    total_route_ports = 0
}
resource "cloudfoundry_user" "u1" {
    name = "test-user1@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u2" {
    name = "test-user2@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u3" {
    name = "test-user3@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u4" {
    name = "test-user4@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u5" {
    name = "test-user5@acme.com"
	password = "password"
}

resource "cloudfoundry_org" "org1" {

    name = "organization-one"
    quota = "${cloudfoundry_org_quota.runaway.id}"
    managers = [ "${cloudfoundry_user.u1.id}", "${cloudfoundry_user.u2.id}" ]
    billing_managers = [ "${cloudfoundry_user.u3.id}", "${cloudfoundry_user.u4.id}" ]
    auditors = [ "${cloudfoundry_user.u5.id}" ]
}
`

const orgResourceUpdate = `

data "cloudfoundry_org_quota" "default" {
  name = "default"
}

resource "cloudfoundry_org_quota" "runaway" {
	name = "runaway_test"
    allow_paid_service_plans = true
    instance_memory = -1
    total_app_instances = -1
    total_memory = 204800
    total_routes = 2000
    total_services = -1
    total_route_ports = 0
}
resource "cloudfoundry_user" "u1" {
    name = "test-user1@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u2" {
    name = "test-user2@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u3" {
    name = "test-user3@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u4" {
    name = "test-user4@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u5" {
    name = "test-user5@acme.com"
	password = "password"
}

resource "cloudfoundry_org" "org1" {
	name = "organization-one-updated"
    quota = "${data.cloudfoundry_org_quota.default.id}"
	managers = [ "${cloudfoundry_user.u1.id}" ]
	billing_managers = [ "${cloudfoundry_user.u2.id}", "${cloudfoundry_user.u3.id}" ]
	auditors = [ "${cloudfoundry_user.u5.id}" ]
}
`

func TestAccResOrg_normal(t *testing.T) {

	refOrg := "cloudfoundry_org.org1"
	refQuotaRunway := "cloudfoundry_org_quota.runaway"
	refQuotaDefault := "data.cloudfoundry_org_quota.default"
	refUserRemoved := "cloudfoundry_user.u4"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckOrgDestroyed("organization-one-updated"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: orgResource,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckOrgExists(refOrg, refQuotaRunway, nil),
						resource.TestCheckResourceAttr(
							refOrg, "name", "organization-one"),
						resource.TestCheckResourceAttr(
							refOrg, "managers.#", "2"),
						resource.TestCheckResourceAttr(
							refOrg, "billing_managers.#", "2"),
						resource.TestCheckResourceAttr(
							refOrg, "auditors.#", "1"),
					),
				},

				resource.TestStep{
					Config: orgResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckOrgExists(refOrg, refQuotaDefault, &refUserRemoved),
						resource.TestCheckResourceAttr(
							refOrg, "name", "organization-one-updated"),
						resource.TestCheckResourceAttr(
							refOrg, "managers.#", "1"),
						resource.TestCheckResourceAttr(
							refOrg, "billing_managers.#", "2"),
						resource.TestCheckResourceAttr(
							refOrg, "auditors.#", "1"),
					),
				},
			},
		})
}

func testAccCheckOrgExists(resOrg, resQuota string, refUserRemoved *string) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resOrg]
		if !ok {
			return fmt.Errorf("org '%s' not found in terraform state", resOrg)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		org, _, err := session.ClientV2.GetOrganization(id)
		if err != nil {
			return err
		}

		if err = assertEquals(attributes, "name", org.Name); err != nil {
			return err
		}
		if err = assertEquals(attributes, "quota", org.QuotaDefinitionGUID); err != nil {
			return err
		}

		rs = s.RootModule().Resources[resQuota]
		if org.QuotaDefinitionGUID != rs.Primary.ID {
			return fmt.Errorf("expected org '%s' to be associated with quota '%s' but it was not", resOrg, resQuota)
		}

		for t, r := range orgRoleMap {
			var users []interface{}
			usersClient, _, err := session.ClientV2.GetOrganizationUsersByRole(r, id)
			if err != nil {
				return err
			}
			users = UsersToIDs(usersClient)
			if err = assertSetEquals(attributes, t, users); err != nil {
				return err
			}
		}

		return testUserRemovedFromOrg(refUserRemoved, id, session, s)
	}
}

func testAccCheckOrgDestroyed(orgname string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)
		orgs, _, err := session.ClientV2.GetOrganizations(ccv2.FilterByName(orgname))
		if err != nil {
			return err
		}
		if len(orgs) > 0 {
			return fmt.Errorf("org with name '%s' still exists in cloud foundry", orgname)
		}
		return nil
	}
}

func testUserRemovedFromOrg(
	refUserRemoved *string,
	orgID string,
	session *managers.Session,
	s *terraform.State) error {

	if refUserRemoved != nil {

		rs, found := s.RootModule().Resources[*refUserRemoved]
		if !found {
			return fmt.Errorf("expected user resource '%s' was not found", *refUserRemoved)
		}

		usersClients, _, err := session.ClientV2.GetOrganizationUsers(orgID)
		if err != nil {
			return err
		}

		found = false
		for _, u := range UsersToIDs(usersClients) {
			if rs.Primary.ID == u {
				found = true
				break
			}
		}
		if found {
			return fmt.Errorf(
				"expected user resource '%s' with if '%s' to be removed from the organization but it was not",
				*refUserRemoved, rs.Primary.ID)
		}
	}
	return nil
}
