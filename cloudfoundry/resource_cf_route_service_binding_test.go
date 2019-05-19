package cloudfoundry

import (
	"bytes"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gopkg.in/src-d/go-git.v4"
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
  url = "file://{{ .CloneDir }}/servicebroker"
  route {
    default_route = "${cloudfoundry_route.basic-auth-broker.id}"
  }
  environment {
    BROKER_CONFIG_PATH = "config.yml"
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
  url = "file://{{ .CloneDir }}/routeserver"
  route {
    default_route = "${cloudfoundry_route.basic-auth-router.id}"
  }
  timeout = 300
}

resource "cloudfoundry_service_instance" "basic-auth" {
  depends_on = [ "cloudfoundry_service_plan_access.basic-auth-access" ]
  name = "basic-auth"
  space = "${data.cloudfoundry_space.space.id}"
  service_plan = "${cloudfoundry_service_broker.basic-auth.service_plans["p-basic-auth/reverse-name"]}"
}

resource "cloudfoundry_route" "php-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "php-app"
}

resource "cloudfoundry_app" "php-app" {
  name = "php-app"
  space = "${data.cloudfoundry_space.space.id}"
  memory = "128"
  disk_quota = "256"
  url = "file://{{ .BaseDir }}/tests/phpapp"
  route {
    default_route = "${cloudfoundry_route.php-app.id}"
  }
  timeout = 300
}
`

const routeBindingResourceCreate = `
resource "cloudfoundry_route_service_binding" "route-bind" {
  service_instance = "${cloudfoundry_service_instance.basic-auth.id}"
  route = "${cloudfoundry_route.php-app.id}"
}
`

const routeBindingResourceUpdate = `
resource "cloudfoundry_route" "php-app-other" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "php-app-other"
}

resource "cloudfoundry_route_service_binding" "route-bind" {
  service_instance = "${cloudfoundry_service_instance.basic-auth.id}"
  route = "${cloudfoundry_route.php-app-other.id}"
}
`

const routeBindingResourceDelete = `
resource "cloudfoundry_route" "php-app-other" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "php-app-other"
}
`

const routeServiceBrokerCfgYml = `---
basic_auth_service_broker:
  route_service_url: "https://basic-auth-router.{{ .Domain }}"
  broker_username: "admin"
  broker_password: "letmein"
`

func TestAccRouteServiceBinding_normal(t *testing.T) {
	dir, err := ioutil.TempDir("", "git-clone")
	if err != nil {
		t.Fatal("unable to create temporary directory")
	}

	url := "https://github.com/mevansam/cf-basic-auth-route-service.git"
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		t.Fatalf("unable to clone repository '%s'", url)
	}

	tpl, _ := template.New("sql").Parse(routeServiceBrokerCfgYml)
	buf := &bytes.Buffer{}
	tpl.Execute(buf, map[string]interface{}{
		"Domain": defaultAppDomain(),
	})
	err = ioutil.WriteFile(dir+"/servicebroker/config.yml", buf.Bytes(), 0666)

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)

	ref := "cloudfoundry_route_service_binding.route-bind"
	tpl, _ = template.New("sql").Parse(routeBindingResourceCommon)
	buf = &bytes.Buffer{}
	tpl.Execute(buf, map[string]interface{}{
		"Domain":   defaultAppDomain(),
		"Org":      orgName,
		"Space":    spaceName,
		"BaseDir":  defaultBaseDir(),
		"CloneDir": dir,
	})
	appURL := fmt.Sprintf("http://php-app.%s", defaultAppDomain())

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: buf.String() + routeBindingResourceCreate,
					Check: resource.ComposeTestCheckFunc(
						checkRouteServiceBindingResource(ref, "cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.php-app"),
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.php-app", true),
						checkAppResponse(appURL, 401),
					),
				},
				resource.TestStep{
					Config: buf.String() + routeBindingResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						checkRouteServiceBindingResource(ref, "cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.php-app-other"),
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.php-app", false),
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.php-app-other", true),
						checkAppResponse(appURL, 200),
					),
				},
				resource.TestStep{
					Config: buf.String() + routeBindingResourceDelete,
					Check: resource.ComposeTestCheckFunc(
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.php-app", false),
						checkRouteServiceBinding("cloudfoundry_service_instance.basic-auth", "cloudfoundry_route.php-app-other", false),
						checkAppResponse(appURL, 200),
					),
				},
			},
		})

	os.RemoveAll(dir)
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
