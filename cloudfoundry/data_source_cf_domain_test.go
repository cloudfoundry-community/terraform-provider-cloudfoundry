package cloudfoundry

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const domainDataResource = `
data "cloudfoundry_domain" "my-domain" {
    name = "%s"
}
`

const privateDomainDataResource = `
resource "cloudfoundry_org" "myorg" {
	name = "myorg-ds-domain"
}

resource "cloudfoundry_domain" "mydomain" {
  sub_domain = "private-ds-domain"
  domain     = "%[1]s"
  org = "${cloudfoundry_org.myorg.id}"
}

data "cloudfoundry_domain" "private" {
  domain = "${cloudfoundry_domain.mydomain.domain}"
  sub_domain = "private-ds-domain"
}
`

func TestAccDataSourceDomain_normal(t *testing.T) {

	domain := strings.Join(strings.Split(defaultAppDomain(), ".")[1:], ".")
	ref := "data.cloudfoundry_domain.my-domain"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(domainDataResource, domain),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							ref, "name", domain),
						resource.TestCheckResourceAttr(
							ref, "domain", strings.SplitN(domain, ".", 2)[1]),
					),
				},
			},
		})
}

func TestAccDataSourceDomain_private(t *testing.T) {
	ref := "data.cloudfoundry_domain.private"
	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(privateDomainDataResource, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(ref, "name", "private-ds-domain."+defaultAppDomain()),
						resource.TestCheckResourceAttr(ref, "sub_domain", "private-ds-domain"),
						resource.TestCheckResourceAttr(ref, "domain", defaultAppDomain()),
					),
				},
			},
		})
}
