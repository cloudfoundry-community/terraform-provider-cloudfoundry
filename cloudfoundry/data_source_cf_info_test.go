package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const infoDataResource = `

data "cloudfoundry_info" "info" {}
`

func TestAccDataSourceInfo_normal(t *testing.T) {

	ref := "data.cloudfoundry_info.info"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: infoDataResource,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							ref, "auth_endpoint", "https://login."+defaultSysDomain()),
						resource.TestCheckResourceAttr(
							ref, "uaa_endpoint", "https://uaa."+defaultSysDomain()),
						resource.TestCheckResourceAttr(
							ref, "logging_endpoint", "wss://loggregator."+defaultSysDomain()+":443"),
						resource.TestCheckResourceAttr(
							ref, "doppler_endpoint", "wss://doppler."+defaultSysDomain()+":443"),
					),
				},
			},
		})
}
