package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const domainResourceShared = `

resource "cloudfoundry_domain" "shared" {
  sub_domain = "dev-res"
  domain = "%s"
}
`

const domainResourceSharedTCP = `

data "cloudfoundry_router_group" "tcp" {
  name = "default-tcp"
}

resource "cloudfoundry_domain" "shared-tcp" {
  sub_domain = "tcp-test-res"
  domain = "%s"
  router_group = "${data.cloudfoundry_router_group.tcp.id}"
}
`

const domainResourcePrivate = `

resource "cloudfoundry_domain" "private" {
	name = "%s.%s"
  org = "%s"
}
`

func TestAccResSharedDomain_normal(t *testing.T) {

	ref := "cloudfoundry_domain.shared"
	domainname := "dev-res." + defaultAppDomain()

	resource.ParallelTest(t,
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
							ref, "sub_domain", "dev-res"),
						resource.TestCheckResourceAttr(
							ref, "domain", defaultAppDomain()),
					),
				},
				resource.TestStep{
					ResourceName:      ref,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}

func TestAccResSharedTCPDomain_normal(t *testing.T) {

	ref := "cloudfoundry_domain.shared-tcp"
	domainname := "tcp-test-res." + defaultAppDomain()

	resource.ParallelTest(t,
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
							ref, "sub_domain", "tcp-test-res"),
						resource.TestCheckResourceAttr(
							ref, "domain", defaultAppDomain()),
						resource.TestCheckResourceAttr(
							ref, "router_type", "tcp"),
					),
				},
			},
		})
}

func TestAccResPrivateDomain_normal(t *testing.T) {

	ref := "cloudfoundry_domain.private"

	domain := "io"
	subDomain := "test-domain-res"

	orgID, _ := defaultTestOrg(t)

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckPrivateDomainDestroy(domain),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(domainResourcePrivate,
						subDomain, domain, orgID),
					Check: resource.ComposeTestCheckFunc(
						checkPrivateDomainExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", subDomain+"."+domain),
						resource.TestCheckResourceAttr(
							ref, "sub_domain", subDomain),
						resource.TestCheckResourceAttr(
							ref, "domain", domain),
						resource.TestCheckResourceAttr(
							ref, "org", orgID),
					),
				},
			},
		})
}

func checkShareDomainExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("domain '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes
		name := attributes["name"]

		dm := session.ClientV2
		domains, _, err := dm.GetSharedDomains(ccv2.FilterByName(name))
		if err != nil {
			return err
		}
		if len(domains) == 0 {
			return NotFound
		}
		if id != domains[0].GUID {
			return fmt.Errorf("expecting domain guid to be '%s' but got '%session'", id, domains[0].GUID)
		}
		return nil
	}
}

func checkPrivateDomainExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("domain '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes
		name := attributes["name"]

		dm := session.ClientV2
		domains, _, err := dm.GetPrivateDomains(ccv2.FilterByName(name))
		if err != nil {
			return err
		}
		if len(domains) == 0 {
			return NotFound
		}
		if id != domains[0].GUID {
			return fmt.Errorf("expecting domain guid to be '%s' but got '%session'", id, domains[0].GUID)
		}
		return nil
	}
}

func testAccCheckSharedDomainDestroy(domainname string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
		domains, _, err := session.ClientV2.GetSharedDomains(ccv2.FilterByName(domainname))
		if err != nil {
			return err
		}
		if len(domains) > 0 {
			return fmt.Errorf("shared domain with name '%s' still exists in cloud foundry", domainname)
		}
		return nil
	}
}

func testAccCheckPrivateDomainDestroy(domainname string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
		domains, _, err := session.ClientV2.GetPrivateDomains(ccv2.FilterByName(domainname))
		if err != nil {
			return err
		}
		if len(domains) > 0 {
			return fmt.Errorf("shared domain with name '%s' still exists in cloud foundry", domainname)
		}
		return nil
	}
}
