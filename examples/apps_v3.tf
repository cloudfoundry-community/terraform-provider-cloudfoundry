resource "cloudfoundry_route" "routes" {  # memo: End result URL is: <hostname>.<domain>
  count = 1
  hostname = "test_app_thanh"
  domain   = data.cloudfoundry_domain.domain_mydomain.id
  space    = data.cloudfoundry_space.space.id
}

resource "cloudfoundry_user_provided_service" "mq" {
  name = "mq_server_v3"
  space = data.cloudfoundry_space.space.id
  credentials = {
    "url" = "mq:#localhost:9000"
    "username" = "tphan"
    "password" = "tphan"
  }
}

resource "cloudfoundry_app" "test_app_thanh" {
    name                    = "ipa-store-test"
    space                   = data.cloudfoundry_space.space.id
    memory                  = 64
    disk_quota              = 2048
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
        service_instance =  resource.cloudfoundry_user_provided_service.mq.id 
    }
}