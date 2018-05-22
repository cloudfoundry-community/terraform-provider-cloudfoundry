package cloudfoundry

import (
	"fmt"
	"strconv"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const ldapUserResource = `

resource "cf_user" "manager1" {
    name = "manager1@acme.com"
    origin = "ldap"
}
`

const userResourceWithGroups = `

resource "cf_user" "admin-service-user" {
    name = "cf-admin"
	password = "qwerty"
	given_name = "Build"
	family_name = "User"
    groups = [ "cloud_controller.admin", "scim.read", "scim.write" ]
}
`

const userResourceWithGroupsUpdate = `

resource "cf_user" "admin-service-user" {
    name = "cf-admin"
	password = "asdfg"
	email = "cf-admin@acme.com"
    groups = [ "cloud_controller.admin", "clients.admin", "uaa.admin", "doppler.firehose" ]
}
`

const userResourceWithEmptyGroup = `

resource "cf_user" "empty-group" {
	name = "jdoe"
	password = "password"
	origin = "uaa"
	given_name = "John"
	family_name = "Doe"
	email = "john.doe@acme.com"
	groups = []
}
`
const userResourceWithEmptyGroupUpdate = `

resource "cf_user" "empty-group" {
	name = "jdoe"
	password = "password"
	origin = "uaa"
	given_name = "John2"
	family_name = "Doe2"
	email = "john.doe@acme.com"
	groups = []
}
`

func TestAccUser_LdapOrigin_normal(t *testing.T) {

	ref := "cf_user.manager1"
	username := "manager1@acme.com"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserDestroy(username),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: ldapUserResource,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", username),
						resource.TestCheckResourceAttr(
							ref, "origin", "ldap"),
						resource.TestCheckResourceAttr(
							ref, "email", username),
					),
				},
			},
		})
}

func TestAccUser_WithGroups_normal(t *testing.T) {

	ref := "cf_user.admin-service-user"
	username := "cf-admin"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserDestroy(username),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: userResourceWithGroups,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", username),
						resource.TestCheckResourceAttr(
							ref, "password", "qwerty"),
						resource.TestCheckResourceAttr(
							ref, "email", username),
						resource.TestCheckResourceAttr(
							ref, "groups.#", "3"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("cloud_controller.admin")),
							"cloud_controller.admin"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("scim.read")),
							"scim.read"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("scim.write")),
							"scim.write"),
					),
				},

				resource.TestStep{
					Config: userResourceWithGroupsUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "cf-admin"),
						resource.TestCheckResourceAttr(
							ref, "password", "asdfg"),
						resource.TestCheckResourceAttr(
							ref, "email", "cf-admin@acme.com"),
						resource.TestCheckResourceAttr(
							ref, "groups.#", "4"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("cloud_controller.admin")),
							"cloud_controller.admin"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("clients.admin")),
							"clients.admin"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("uaa.admin")),
							"uaa.admin"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("doppler.firehose")),
							"doppler.firehose"),
					),
				},
			},
		})
}

func TestAccUser_EmptyGroups_normal(t *testing.T) {

	ref := "cf_user.empty-group"
	username := "jdoe"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserDestroy(username),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: userResourceWithEmptyGroup,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", username),
						resource.TestCheckResourceAttr(
							ref, "origin", "uaa"),
						resource.TestCheckResourceAttr(
							ref, "email", "john.doe@acme.com"),
					),
				},

				resource.TestStep{
					Config: userResourceWithEmptyGroupUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", username),
						resource.TestCheckResourceAttr(
							ref, "origin", "uaa"),
						resource.TestCheckResourceAttr(
							ref, "email", "john.doe@acme.com"),
					),
				},
			},
		})
}

func testAccCheckUserExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("user '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		um := session.UserManager()
		user, err := um.GetUser(id)
		if err != nil {
			return err
		}

		session.Log.DebugMessage(
			"retrieved user for resource '%s' with id '%s': %# v",
			resource, id, user)

		if err = assertEquals(attributes, "name", user.Username); err != nil {
			return err
		}
		if err = assertEquals(attributes, "origin", user.Origin); err != nil {
			return err
		}
		if err = assertEquals(attributes, "given_name", user.Name.GivenName); err != nil {
			return err
		}
		if err = assertEquals(attributes, "family_name", user.Name.FamilyName); err != nil {
			return err
		}
		if err = assertEquals(attributes, "email", user.Emails[0].Value); err != nil {
			return err
		}

		var groups []interface{}
		for _, g := range user.Groups {
			if !um.IsDefaultGroup(g.Display) {
				groups = append(groups, g.Display)
			}
		}
		err = assertSetEquals(attributes, "groups", groups)
		return err
	}
}

func testAccCheckUserDestroy(username string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*cfapi.Session)
		um := session.UserManager()
		if _, err := um.FindByUsername(username); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}
		return fmt.Errorf("user with username '%s' still exists in cloud foundry", username)
	}
}
