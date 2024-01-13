package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const appDataSource = `

data "cf_org" "org" {
	name = "%s"
}
data "cf_space" "space" {
	name = "%s"
	org = "${data.cf_org.org.id}"
}

resource "cf_app" "dummy-app" {
	name = "data_source_cf_app_tests_%s"
	space = "${data.cf_space.space.id}"
	memory = 1024
	disk_quota = 1024
	instances = 1
    health_check_type          = "port"
	health_check_timeout       = 180

	stack = "cflinuxfs4"

	buildpacks = [
		"staticfile_buildpack",
		"nodejs_buildpack"
		
	]

	enable_ssh = true

	environment = {
		"UPDATED" = "true"
		"APPNAME" = "cf_app_test"
	}
	
	path = "%s"
}

data "cf_app" "dummy-app" {
	name_or_id = cf_app.dummy-app.id
	space = "${data.cf_space.space.id}"
	depends_on = [ cf_app.dummy-app ]
}

output "dummy"{
	value = data.cf_app.dummy-app.id
}
`

func TestAccDataSourceApp_normal(t *testing.T) {
	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)

	ref := "data.cf_app.dummy-app"
	appSuffix := uuid.New()

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appDataSource, orgName, spaceName, appSuffix, "https://int.repositories.cloud.sap/artifactory/build-milestones/com/sap/npm/cf-route-service-dummy-app/1.0.1/cf-route-service-dummy-app-1.0.1-bundle.tar.gz"),
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							_, ok := s.RootModule().Resources[ref]
							if !ok {
								return fmt.Errorf("app '%s' not found in terraform state", ref)
							}
							return nil
						},
						resource.TestCheckResourceAttr(ref, "name", fmt.Sprintf("data_source_cf_app_tests_%s", appSuffix)),
						resource.TestCheckResourceAttr(ref, "instances", "1"),
						resource.TestCheckResourceAttr(ref, "memory", "1024"),
						resource.TestCheckResourceAttr(ref, "disk_quota", "1024"),
						resource.TestCheckResourceAttr(ref, "stack", "cflinuxfs4"),
						resource.TestCheckResourceAttr(ref, "buildpack", "staticfile_buildpack"),
						resource.TestCheckResourceAttr(ref, "command", "npm start"),
						resource.TestCheckResourceAttr(ref, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(ref, "state", "STARTED"),
						resource.TestCheckResourceAttr(ref, "health_check_type", "port"),
						resource.TestCheckResourceAttr(ref, "health_check_timeout", "180"),
						resource.TestCheckResourceAttr(ref, "buildpacks.0", "staticfile_buildpack"),
						resource.TestCheckResourceAttr(ref, "buildpacks.1", "nodejs_buildpack"),
						resource.TestCheckResourceAttr(ref, "environment.%", "2"),
						resource.TestCheckResourceAttr(ref, "environment.UPDATED", "true"),
						resource.TestCheckResourceAttr(ref, "environment.APPNAME", "cf_app_test"),
					),
				},
			},
		})
}
