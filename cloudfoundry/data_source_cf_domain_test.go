package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

const domainDataResource = `

data "cf_domain" "tcp" {
    sub_domain = "tcp"
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
