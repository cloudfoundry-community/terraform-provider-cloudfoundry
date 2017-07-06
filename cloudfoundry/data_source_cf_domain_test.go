package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

const domainDataResource = `

data "cf_domain" "apps" {
    sub_domain = "local"
}
`

func TestAccDataSourceDomain_normal(t *testing.T) {

	ref := "data.cf_domain.apps"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: domainDataResource,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							ref, "name", "local.pcfdev.io"),
						resource.TestCheckResourceAttr(
							ref, "sub_domain", "local"),
						resource.TestCheckResourceAttr(
							ref, "domain", "pcfdev.io"),
					),
				},
			},
		})
}
