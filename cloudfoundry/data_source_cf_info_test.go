package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

const infoDataResource = `

data "cf_info" "info" {}
`

func TestAccDataSourceInfo_normal(t *testing.T) {

	ref := "data.cf_info.info"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: infoDataResource,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							ref, "api_endpoint", apiURL()),
						resource.TestCheckResourceAttr(
							ref, "auth_endpoint", "https://login."+defaultDomain()),
						resource.TestCheckResourceAttr(
							ref, "uaa_endpoint", "https://uaa."+defaultDomain()),
						resource.TestCheckResourceAttr(
							ref, "logging_endpoint", "wss://loggregator."+defaultDomain()+":443"),
						resource.TestCheckResourceAttr(
							ref, "doppler_endpoint", "wss://doppler."+defaultDomain()+":443"),
					),
				},
			},
		})
}
