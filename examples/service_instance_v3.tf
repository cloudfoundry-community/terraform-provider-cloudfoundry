## Grab a Cloud Foundry application-logs details
data "cloudfoundry_service" "application-logs" {
  name = "application-logs"
}

# ## Create a Cloud Foundry application-logs service
# resource "cloudfoundry_service_instance_v3" "application-logs" {
#   name             = "app-log"
#   space            = "${data.cloudfoundry_space_v3.space_v3.id}"
#   service_plan     = "${data.cloudfoundry_service_offering_v3.application-logs.service_plans["lite"]}"
#   #tags            = var.cf_si_tags 
# }

## Service output
# output "application-logs-id" {
#   value = "${cloudfoundry_service_instance_v3.application-logs.id}"
# }
