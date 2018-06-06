package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

const domainDataResource = `
data "cf_domain" "tcp" {
    sub_domain = "tcp"
}
`

const privateDomainDataResource = `
resource "cf_org" "myorg" {
	name = "myorg"
}

resource "cf_domain" "mydomain" {
  sub_domain = "private"
  domain     = "%[1]s"
  org = "${cf_org.myorg.id}"
}

data "cf_domain" "private" {
  domain = "${cf_domain.mydomain.domain}"
  sub_domain = "private"
}
`

func TestAccDataSourceDomain_normal(t *testing.T) {

	ref := "data.cf_domain.tcp"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: domainDataResource,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							ref, "name", "tcp."+defaultAppDomain()),
						resource.TestCheckResourceAttr(
							ref, "sub_domain", "tcp"),
						resource.TestCheckResourceAttr(
							ref, "domain", defaultAppDomain()),
					),
				},
			},
		})
}

func TestAccDataSourceDomain_private(t *testing.T) {
	ref := "data.cf_domain.private"
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
