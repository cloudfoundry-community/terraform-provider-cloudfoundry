package cloudfoundry

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const routeBindingResourceCommon = `
data "cloudfoundry_domain" "local" {
  name = "{{ .Domain }}"
}

data "cloudfoundry_org" "org" {
  name = "{{ .Org }}"
}

data "cloudfoundry_space" "space" {
  name = "{{ .Space }}"
  org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_route" "basic-auth-broker" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "basic-auth-broker"
}

resource "cloudfoundry_app" "basic-auth-broker" {
  name = "basic-auth-broker"
  space = "${data.cloudfoundry_space.space.id}"
  memory = "128"
  disk_quota = "256"
  path = "{{ .BrokerPath }}"
  buildpack = "binary_buildpack"
  command = "./servicebroker"
  routes {
    route = "${cloudfoundry_route.basic-auth-broker.id}"
  }
  environment = {
    BROKER_CONFIG_PATH = "config.yml"
	ROUTE_SERVICE_URL = "https://basic-auth-router.{{ .Domain }}"
  }
  timeout = 300
}

resource "cloudfoundry_service_broker" "basic-auth" {
  depends_on = [ "cloudfoundry_app.basic-auth-broker" ]
  name = "basic-auth"
  url = "https://basic-auth-broker.{{ .Domain }}"
  username = "admin"
  password = "letmein"
}

resource "cloudfoundry_service_plan_access" "basic-auth-access" {
  plan = "${cloudfoundry_service_broker.basic-auth.service_plans["p-basic-auth/reverse-name"]}"
  public = true
}


resource "cloudfoundry_route" "basic-auth-router" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "basic-auth-router"
}

resource "cloudfoundry_app" "basic-auth-router" {
  name = "basic-auth-router"
  space = "${data.cloudfoundry_space.space.id}"
  memory = "128"
  disk_quota = "256"
  path = "{{ .ServerPath }}"
  buildpack = "binary_buildpack"
  command = "./routeserver"
  routes {
	route = "${cloudfoundry_route.basic-auth-router.id}"
  }
  timeout = 300
}

resource "cloudfoundry_service_instance" "basic-auth" {
  depends_on = [ "cloudfoundry_service_plan_access.basic-auth-access" ]
  name = "basic-auth"
  space = "${data.cloudfoundry_space.space.id}"
  service_plan = "${cloudfoundry_service_broker.basic-auth.service_plans["p-basic-auth/reverse-name"]}"
}

resource "cloudfoundry_route" "dummy-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "dummy-app"
}

resource "cloudfoundry_app" "dummy-app" {
  name = "dummy-app"
  buildpack = "binary_buildpack"
  space = "${data.cloudfoundry_space.space.id}"
  memory = "128"
  disk_quota = "256"
  path = "{{ .DummyAppPath }}"
  routes {
    route = "${cloudfoundry_route.dummy-app.id}"
  }
  timeout = 300
}
`

const routeBindingResourceCreate = `
resource "cloudfoundry_route_service_binding" "route-bind" {
  service_instance = "${cloudfoundry_service_instance.basic-auth.id}"
  route = "${cloudfoundry_route.dummy-app.id}"
}
`

const routeBindingResourceUpdate = `
resource "cloudfoundry_route" "dummy-app-other" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "dummy-app-other"
}

resource "cloudfoundry_route_service_binding" "route-bind" {
  service_instance = "${cloudfoundry_service_instance.basic-auth.id}"
  route = "${cloudfoundry_route.dummy-app-other.id}"
}
`

const routeBindingResourceDelete = `
resource "cloudfoundry_route" "dummy-app-other" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "dummy-app-other"
}
`

func TestAccResRouteServiceBinding_normal(t *testing.T) {
	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)

	ref := "cloudfoundry_route_service_binding.route-bind"
	tpl, _ := template.New("sql").Parse(routeBindingResourceCommon)
	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, map[string]interface{}{
		"Domain":       defaultAppDomain(),
		"Org":          orgName,
		"Space":        spaceName,
		"BaseDir":      defaultBaseDir(),
		"BrokerPath":   asset("route-service-broker.zip"),
		"ServerPath":   asset("route-service-server.zip"),
		"DummyAppPath": asset("dummy-app.zip"),
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	appURL := fmt.Sprintf("http://dummy-app.%s", defaultAppDomain())

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: buf.String() + routeBindingResourceCreate,
					Check: resource.ComposeTestCheckFunc(
						checkRouteServiceBindingResource(ref, "cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.dummy-app"),
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.dummy-app", true),
						checkAppResponse(appURL, 401),
					),
				},
				resource.TestStep{
					Config: buf.String() + routeBindingResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						checkRouteServiceBindingResource(ref, "cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.dummy-app-other"),
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.dummy-app", false),
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.dummy-app-other", true),
						checkAppResponse(appURL, 200),
					),
				},
				resource.TestStep{
					Config: buf.String() + routeBindingResourceDelete,
					Check: resource.ComposeTestCheckFunc(
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.dummy-app", false),
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.dummy-app-other", false),
						checkAppResponse(appURL, 200),
					),
				},
			},
		})
}

func checkAppResponse(url string, code int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resp, err := testSession().HttpClient.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != code {
			// on cf without valid certif it has 502 error
			if resp.StatusCode == 502 {
				fmt.Printf("Warning, looks like is in self signed router ssl env so let's pass 502 instead of %d\n", code)
				return nil
			}
			return fmt.Errorf("invalid status code '%d', expected '%d'", resp.StatusCode, code)
		}
		return nil
	}
}

func checkRouteServiceBinding(serviceName, routeName string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
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
		routes, _, err := session.ClientV2.GetServiceBindingRoutes(serviceID)
		if err != nil {
			return err
		}
		found := false
		for _, route := range routes {
			if route.GUID == routeID {
				found = true
				break
			}
		}
		if !found && exists {
			return fmt.Errorf("unable to find route '%s(%s)' binding to service '%s(%s)'", serviceName, serviceID, routeName, routeID)
		}

		if found && !exists {
			return fmt.Errorf("route '%s(%s)' binding to service '%s(%s)' not deleted as it ought to be", serviceName, serviceID, routeName, routeID)
		}

		return nil
	}
}

func checkRouteServiceBindingResource(resource, serviceName, routeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
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

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("route_service_binding '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		if id != fmt.Sprintf("%s/%s", serviceID, routeID) {
			return fmt.Errorf("unexpected route_service_binding resource identifier '%s' mismatch with '%s/%s'", id, serviceID, routeID)
		}

		return nil
	}
}
