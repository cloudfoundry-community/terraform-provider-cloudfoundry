package cloudfoundry

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
	"net/http"
	"testing"
	"text/template"
)

const routeBindingResourceCommon = `
data "cf_domain" "local" {
    name = "{{ .Domain }}"
}

data "cf_org" "org" {
  name = "pcfdev-org"
}

data "cf_space" "space" {
  name = "pcfdev-space"
  org = "${data.cf_org.org.id}"
}

resource "cf_route" "basic-auth-broker" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "basic-auth-broker"
}

resource "cf_app" "basic-auth-broker" {
	name = "basic-auth-broker"
	space = "${data.cf_space.space.id}"
	memory = "128"
	disk_quota = "256"
  url = "file://{{ .BaseDir }}/tests/servicebroker"
	route {
		default_route = "${cf_route.basic-auth-broker.id}"
	}
  environment {
    BROKER_CONFIG_PATH = "config.yml"
  }
  timeout = 300
}

resource "cf_service_broker" "basic-auth" {
  depends_on = [ "cf_app.basic-auth-broker" ]
	name = "basic-auth"
  url = "https://basic-auth-broker.{{ .Domain }}"
	username = "admin"
	password = "letmein"
}

resource "cf_service_plan_access" "basic-auth-access" {
	plan = "${cf_service_broker.basic-auth.service_plans["p-basic-auth/reverse-name"]}"
	public = true
}


resource "cf_route" "basic-auth-app" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "basic-auth-app"
}

resource "cf_app" "basic-auth-app" {
	name = "basic-auth-app"
	space = "${data.cf_space.space.id}"
	memory = "128"
	disk_quota = "256"
  url = "file://{{ .BaseDir }}/tests/routeserver"
	route {
		default_route = "${cf_route.basic-auth-app.id}"
	}
  timeout = 300
}

resource "cf_service_instance" "basic-auth" {
  depends_on = [ "cf_service_plan_access.basic-auth-access" ]
	name = "basic-auth"
  space = "${data.cf_space.space.id}"
  service_plan = "${cf_service_broker.basic-auth.service_plans["p-basic-auth/reverse-name"]}"
}

resource "cf_route" "php-app" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "php-app"
}

resource "cf_app" "php-app" {
	name = "php-app"
	space = "${data.cf_space.space.id}"
	memory = "128"
	disk_quota = "256"
  url = "file://{{ .BaseDir }}/tests/phpapp"
	route {
		default_route = "${cf_route.php-app.id}"
	}
  timeout = 300
}


`

const routeBindingResourceCreate = `
resource "cf_route_binding" "route-bind" {
  service_instance = "${cf_service_instance.basic-auth.id}"
  route = "${cf_route.php-app.id}"
}
`

const routeBindingResourceUpdate = `
resource "cf_route_binding" "route-bind" {
  service_instance = "${cf_service_instance.basic-auth.id}"
  route = "${cf_route.php-app.id}"
  json_params = <<JSON
  {
    "key1" : "value1",
    "key2" : "value2"
  }
  JSON
}
`

func TestAccRouteBinding_normal(t *testing.T) {
	ref := "cf_route_binding.route-bind"
	tpl, _ := template.New("sql").Parse(routeBindingResourceCommon)
	buf := &bytes.Buffer{}
	tpl.Execute(buf, map[string]interface{}{
		"Domain":  defaultAppDomain(),
		"BaseDir": defaultBaseDir(),
	})
	appUrl := fmt.Sprintf("http://php-app.%s", defaultAppDomain())

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: buf.String() + routeBindingResourceCreate,
					Check: resource.ComposeTestCheckFunc(
						checkRouteBinding(ref, "cf_service_instance.basic-auth", "cf_route.php-app", true),
						checkAppResponse(appUrl, 401),
					),
				},
				resource.TestStep{
					Config: buf.String() + routeBindingResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						checkRouteBinding(ref, "cf_service_instance.basic-auth", "cf_route.php-app", true),
						checkAppResponse(appUrl, 401),
					),
				},
				resource.TestStep{
					Config: buf.String(),
					Check: resource.ComposeTestCheckFunc(
						checkRouteBinding("", "cf_service_instance.basic-auth", "cf_route.php-app", false),
						checkAppResponse(appUrl, 200),
					),
				},
			},
		})
}

func checkAppResponse(url string, code int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != code {
			return fmt.Errorf("invalid status code '%d', expected '%d'", resp.StatusCode, code)
		}
		return nil
	}
}

func checkRouteBinding(resource, serviceName, routeName string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*cfapi.Session)

		service, okService := s.RootModule().Resources[serviceName]
		route, okRoute := s.RootModule().Resources[routeName]

		if !okService {
			return fmt.Errorf("service '%s' not found in terraform state", serviceName)
		}
		if !okRoute {
			return fmt.Errorf("route '%s' not found in terraform state", routeName)
		}

		serviceID := service.Primary.ID
		routeID := route.Primary.ID

		sm := session.ServiceManager()
		found, err := sm.HasRouteBinding(serviceID, routeID)
		if err != nil {
			return err
		}

		if !found && exists {
			return fmt.Errorf("unable to find route '%s(%s)' binding to service '%s(%s)'", serviceName, serviceID, routeName, routeID)
		}

		if found && !exists {
			return fmt.Errorf("route '%s(%s)' binding to service '%s(%s)' not deleted as it ought to be", serviceName, serviceID, routeName, routeID)
		}

		if len(resource) > 0 {
			rs, ok := s.RootModule().Resources[resource]
			if !ok {
				return fmt.Errorf("route_binding '%s' not found in terraform state", resource)
			}
			session.Log.DebugMessage("terraform state for resource '%s': %# v", resource, rs)

			id := rs.Primary.ID

			if exists && id != fmt.Sprintf("%s/%s", serviceID, routeID) {
				return fmt.Errorf("unexpected route_binding resource identifier '%s' mismatch with '%s/%s'", id, serviceID, routeID)
			}
		}
		return nil
	}
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
