package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const appDataSource = `

data "cloudfoundry_org" "org" {
	name = "%s"
}
data "cloudfoundry_space" "space" {
	name = "%s"
	org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_app" "dummy-app" {
	name = "%s"
	space = "${data.cloudfoundry_space.space.id}"

	instances = 1
	health_check_type          = "port"
	health_check_timeout       = 180

	stack = "cflinuxfs4"

	buildpacks = [
		"staticfile_buildpack",
		"binary_buildpack"
		
	]

	enable_ssh = true

	environment = {
		"UPDATED" = "true"
		"APPNAME" = "cloudfoundry_app_test"
	}
	
	path = "%s"
}

data "cloudfoundry_app" "dummy-app-by-id" {
	name_or_id = cloudfoundry_app.dummy-app.id
	space = "${data.cloudfoundry_space.space.id}"
	depends_on = [ cloudfoundry_app.dummy-app ]
}

data "cloudfoundry_app" "dummy-app-by-name" {
	name_or_id = cloudfoundry_app.dummy-app.name
	space = "${data.cloudfoundry_space.space.id}"
	depends_on = [ cloudfoundry_app.dummy-app ]
}

output "dummy"{
	value = [data.cloudfoundry_app.dummy-app-by-id.id, data.cloudfoundry_app.dummy-app-by-name.id]
}
`

func TestAccDataSourceApp_normal(t *testing.T) {
	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)

	refs := []string{"data.cloudfoundry_app.dummy-app-by-id", "data.cloudfoundry_app.dummy-app-by-name"}

	appName := "BindAppName-%s"
	appName = fmt.Sprintf(appName, uuid.New())

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appDataSource, orgName, spaceName, appName, "https://raw.githubusercontent.com/cloudfoundry-community/terraform-provider-cloudfoundry/main/tests/cf-acceptance-tests/assets/dummy-app.zip"),
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							for _, ref := range refs {
								_, ok := s.RootModule().Resources[ref]
								if !ok {
									return fmt.Errorf("app '%s' not found in terraform state", ref)
								}
							}
							return nil
						},
						resource.TestCheckResourceAttr(refs[0], "name", appName),
						resource.TestCheckResourceAttr(refs[0], "instances", "1"),
						resource.TestCheckResourceAttr(refs[0], "memory", "1024"),
						resource.TestCheckResourceAttr(refs[0], "disk_quota", "1024"),
						resource.TestCheckResourceAttr(refs[0], "stack", "cflinuxfs4"),
						resource.TestCheckResourceAttr(refs[0], "buildpack", "staticfile_buildpack"),
						resource.TestCheckResourceAttr(refs[0], "command", "./app"),
						resource.TestCheckResourceAttr(refs[0], "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refs[0], "state", "STARTED"),
						resource.TestCheckResourceAttr(refs[0], "health_check_type", "port"),
						resource.TestCheckResourceAttr(refs[0], "health_check_timeout", "180"),
						resource.TestCheckResourceAttr(refs[0], "buildpacks.0", "staticfile_buildpack"),
						resource.TestCheckResourceAttr(refs[0], "buildpacks.1", "binary_buildpack"),
						resource.TestCheckResourceAttr(refs[0], "environment.%", "2"),
						resource.TestCheckResourceAttr(refs[0], "environment.UPDATED", "true"),
						resource.TestCheckResourceAttr(refs[0], "environment.APPNAME", "cloudfoundry_app_test"),
						resource.TestCheckResourceAttr(refs[1], "name", appName),
						resource.TestCheckResourceAttr(refs[1], "instances", "1"),
						resource.TestCheckResourceAttr(refs[1], "memory", "1024"),
						resource.TestCheckResourceAttr(refs[1], "disk_quota", "1024"),
						resource.TestCheckResourceAttr(refs[1], "stack", "cflinuxfs4"),
						resource.TestCheckResourceAttr(refs[1], "buildpack", "staticfile_buildpack"),
						resource.TestCheckResourceAttr(refs[1], "command", "./app"),
						resource.TestCheckResourceAttr(refs[1], "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refs[1], "state", "STARTED"),
						resource.TestCheckResourceAttr(refs[1], "health_check_type", "port"),
						resource.TestCheckResourceAttr(refs[1], "health_check_timeout", "180"),
						resource.TestCheckResourceAttr(refs[1], "buildpacks.0", "staticfile_buildpack"),
						resource.TestCheckResourceAttr(refs[1], "buildpacks.1", "binary_buildpack"),
						resource.TestCheckResourceAttr(refs[1], "environment.%", "2"),
						resource.TestCheckResourceAttr(refs[1], "environment.UPDATED", "true"),
						resource.TestCheckResourceAttr(refs[1], "environment.APPNAME", "cloudfoundry_app_test"),
					),
				},
			},
		})
}
