package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const orgUsersResource = `
resource "cloudfoundry_user" "tl" {
	name = "teamlead@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "dev1" {
    name = "developer1@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "dev2" {
    name = "developer2@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "dev3" {
    name = "developer3@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "adr" {
    name = "auditor@acme.com"
	password = "password"
}

resource "cloudfoundry_org_users" "org_users1" {
	org = "%s"
    managers = [
        "${cloudfoundry_user.tl.id}"
    ]
    billing_managers = [
        "${cloudfoundry_user.tl.id}",
        "${cloudfoundry_user.dev1.id}",
		"${cloudfoundry_user.dev2.id}"
    ]
    auditors = [
        "${cloudfoundry_user.adr.id}",
		"${cloudfoundry_user.dev3.id}",
        "username@acme.com"
    ]
}
`

const orgUsersResourceUpdate = `
resource "cloudfoundry_user" "dev1" {
    name = "developer1@acme.com"
	password = "password"
}

resource "cloudfoundry_org_users" "org_users1" {
	org = "%s"
	managers = []
    billing_managers = [
        "${cloudfoundry_user.dev1.id}",
    ]
	auditors = []
}
`

const orgUsersResourceForce = `
resource "cloudfoundry_user" "dev1" {
    name = "developer1@acme.com"
	password = "password"
}

resource "cloudfoundry_org_users" "org_users1" {
	org = "%s"
    billing_managers = [
        "${cloudfoundry_user.dev1.id}",
    ]
	force = true
}
`

func TestAccResOrgUsers_normal(t *testing.T) {
	ref := "cloudfoundry_org_users.org_users1"
	orgId, _ := defaultTestOrg(t)
	usersMap := make(map[string][]ccv2.User)

	sessions := testSession()
	user, err := sessions.ClientUAA.CreateUser("username@acme.com", "paasw0rd", "uaa")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer sessions.ClientUAA.DeleteUser(user.ID)
	err = addOrNothingUserInOrgBySpace(sessions, orgId, user.ID, false)
	if err != nil {
		t.Fatal(err.Error())
	}

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(orgUsersResource, orgId),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckOrgUsersExists(ref, &usersMap),
						resource.TestCheckResourceAttr(
							ref, "managers.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "billing_managers.#", "3"),
						resource.TestCheckResourceAttr(
							ref, "auditors.#", "3"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(orgUsersResourceUpdate, orgId),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckOrgUsersExists(ref, &usersMap),
						testAccCheckMapUserInside("developer1@acme.com", "billing_managers", &usersMap),
						resource.TestCheckResourceAttr(
							ref, "managers.#", "0"),
						resource.TestCheckResourceAttr(
							ref, "billing_managers.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "auditors.#", "0"),
					),
				},
			},
		})
}

func TestAccResOrgUsers_force(t *testing.T) {
	ref := "cloudfoundry_org_users.org_users1"
	orgId, _ := defaultTestOrg(t)
	usersMap := make(map[string][]ccv2.User)

	sessions := testSession()
	user, err := sessions.ClientUAA.CreateUser("test-acc-force@acme.com", "paasw0rd", "uaa")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer sessions.ClientUAA.DeleteUser(user.ID)

	_, err = sessions.ClientV2.UpdateOrganizationManager(orgId, user.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(orgUsersResourceForce, orgId),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckOrgUsersExists(ref, &usersMap),
						testAccCheckMapUsersNumber(1, &usersMap),
						testAccCheckMapUserInside("developer1@acme.com", "billing_managers", &usersMap),
						resource.TestCheckResourceAttr(
							ref, "billing_managers.#", "1"),
					),
				},
			},
		})
}

func testAccCheckOrgUsersExists(resource string, users *map[string][]ccv2.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		attributes := rs.Primary.Attributes

		sm := session.ClientV2
		usersMap := make(map[string][]ccv2.User)
		for t, r := range orgRoleMap {
			users, _, err := sm.GetOrganizationUsersByRole(r, attributes["org"])
			if err != nil {
				return err
			}
			usersMap[t] = users
		}
		if len(usersMap) == 0 {
			return fmt.Errorf("No org users found")
		}
		*users = usersMap
		return nil
	}
}
