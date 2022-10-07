resource "cloudfoundry_route" "routes" {  # memo: End result URL is: <hostname>.<domain>
  count = 1
  hostname = "test_app_thanh"
  domain   = data.cloudfoundry_domain_v3.domain_mydomain.id
  space    = data.cloudfoundry_space_v3.space_v3.id
}

resource "cloudfoundry_app_v3" "test_app_thanh" {
    name                    = "ipa-store-test"
    space                   = data.cloudfoundry_space_v3.space_v3.id
    memory                  = 128
    timeout                 = 120
    path                    = "/home/tphan/SAPDevelop/v3_migration/frontend/target/com.sap.ipa.store/dist.zip"
    strategy                = "standard"
    instances = 1

    dynamic "routes" {
        for_each = cloudfoundry_route.routes
        iterator = route
        content {
            route = route.value.id
        }
    }

    service_binding {
        params_json = "{\"credential-type\":\"x509\"}"
        service_instance =  data.cloudfoundry_service_instance.xsuaa.id 
    }
}

resource "cloudfoundry_app" "test_app_thanh_v2" {
    name                    = "ipa-store-test-cfv2"
    space                   = data.cloudfoundry_space_v3.space_v3.id
    memory                  = 128
    timeout                 = 120
    path                    = "/home/tphan/SAPDevelop/v3_migration/frontend/target/com.sap.ipa.store/dist.zip"
    strategy                = "standard"
    instances = 1

    dynamic "routes" {
        for_each = cloudfoundry_route.routes
        iterator = route
        content {
            route = route.value.id
        }
    }

    service_binding {
        params_json = "{\"credential-type\":\"x509\"}"
        service_instance =  data.cloudfoundry_service_instance.xsuaa.id 
    }
}