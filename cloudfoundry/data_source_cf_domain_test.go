package cloudfoundry

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

const domainDataResource = `
data "cloudfoundry_domain" "tcp" {
    sub_domain = "tcp"
}
`

const privateDomainDataResource = `
resource "cloudfoundry_org" "myorg" {
	name = "myorg"
}

resource "cloudfoundry_domain" "mydomain" {
  sub_domain = "private"
  domain     = "%[1]s"
  org = "${cloudfoundry_org.myorg.id}"
}

data "cloudfoundry_domain" "private" {
  domain = "${cloudfoundry_domain.mydomain.domain}"
  sub_domain = "private"
}
`

func TestAccDataSourceDomain_normal(t *testing.T) {

	domain := strings.Join(strings.Split(defaultAppDomain(), ".")[1:], ".")
	ref := "data.cloudfoundry_domain.tcp"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: domainDataResource,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							ref, "name", "tcp."+domain),
						resource.TestCheckResourceAttr(
							ref, "sub_domain", "tcp"),
						resource.TestCheckResourceAttr(
							ref, "domain", domain),
					),
				},
			},
		})
}

func TestAccDataSourceDomain_private(t *testing.T) {
	ref := "data.cloudfoundry_domain.private"
	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(privateDomainDataResource, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(ref, "name", "private."+defaultAppDomain()),
						resource.TestCheckResourceAttr(ref, "sub_domain", "private"),
						resource.TestCheckResourceAttr(ref, "domain", defaultAppDomain()),
					),
				},
			},
		})
}
