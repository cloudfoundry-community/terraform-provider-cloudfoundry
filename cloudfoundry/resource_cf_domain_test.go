package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

const domainResourceShared = `

resource "cloudfoundry_domain" "shared" {
    sub_domain = "dev"
	domain = "%s"
}
`

const domainResourceSharedTCP = `

data "cloudfoundry_router_group" "tcp" {
    name = "default-tcp"
}

resource "cloudfoundry_domain" "shared-tcp" {
    sub_domain = "tcp-test"
	domain = "%s"
	router_group_id = "${data.cloudfoundry_router_group.tcp.id}"
}
`

const domainResourcePrivate = `

resource "cloudfoundry_domain" "private" {
    name = "pcfdev-org.io"
	org_id = "%s"
}
`

func TestAccSharedDomain_normal(t *testing.T) {

	ref := "cloudfoundry_domain.shared"
	domainname := "dev." + defaultAppDomain()

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSharedDomainDestroy(domainname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(domainResourceShared, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						checkShareDomainExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", domainname),
						resource.TestCheckResourceAttr(
							ref, "sub_domain", "dev"),
						resource.TestCheckResourceAttr(
							ref, "domain", defaultAppDomain()),
					),
				},
			},
		})
}

func TestAccSharedTCPDomain_normal(t *testing.T) {

	ref := "cloudfoundry_domain.shared-tcp"
	domainname := "tcp-test." + defaultAppDomain()

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSharedDomainDestroy(domainname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(domainResourceSharedTCP, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						checkShareDomainExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", domainname),
						resource.TestCheckResourceAttr(
							ref, "sub_domain", "tcp-test"),
						resource.TestCheckResourceAttr(
							ref, "domain", defaultAppDomain()),
						resource.TestCheckResourceAttr(
							ref, "router_type", "tcp"),
					),
				},
			},
		})
}

func TestAccPrivateDomain_normal(t *testing.T) {

	ref := "cloudfoundry_domain.private"
	domainname := "pcfdev-org.io"
	orgID := defaultPcfDevOrgID()

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckPrivateDomainDestroy(domainname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(domainResourcePrivate, orgID),
					Check: resource.ComposeTestCheckFunc(
						checkPrivateDomainExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "pcfdev-org.io"),
						resource.TestCheckResourceAttr(
							ref, "sub_domain", "pcfdev-org"),
						resource.TestCheckResourceAttr(
							ref, "domain", "io"),
						resource.TestCheckResourceAttr(
							ref, "org_id", orgID),
					),
				},
			},
		})
}

func checkShareDomainExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("domain '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes
		name := attributes["name"]

		dm := session.DomainManager()
		domainFields, err := dm.FindSharedByName(name)
		if err != nil {
			return err
		}

		if id != domainFields.GUID {
			return fmt.Errorf("expecting domain guid to be '%s' but got '%session'", id, domainFields.GUID)
		}
		return nil
	}
}

func checkPrivateDomainExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("domain '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes
		name := attributes["name"]

		dm := session.DomainManager()
		domainFields, err := dm.FindPrivateByName(name)
		if err != nil {
			return err
		}

		if id != domainFields.GUID {
			return fmt.Errorf("expecting domain guid to be '%s' but got '%session'", id, domainFields.GUID)
		}
		if err := assertEquals(attributes, "org_id", domainFields.OwningOrganizationGUID); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckSharedDomainDestroy(domainname string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)
		_, err := session.DomainManager().FindSharedByName(domainname)
		switch err.(type) {
		case *errors.ModelNotFoundError:
			return nil
		}
		return fmt.Errorf("shared domain with name '%s' still exists in cloud foundry", domainname)
	}
}

func testAccCheckPrivateDomainDestroy(domainname string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*cfapi.Session)
		if _, err := session.DomainManager().FindPrivateByName(domainname); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}
		return fmt.Errorf("domain with name '%s' still exists in cloud foundry", domainname)
	}
}
