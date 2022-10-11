# data "cloudfoundry_service_instance" "xsuaa" {
#   name_or_id = "e9532a3d-f7b8-44dd-a9d5-9a14ce3ca596"
#   space      = data.cloudfoundry_space_v3.space_v3.id
# }

# resource "cloudfoundry_service_instance" "service_instance" {
#   name         = var.service_instance_name
#   space        = var.cf_space_id
#   service_plan = data.cloudfoundry_service.cf_backing_service.service_plans[var.service_plan]
#   json_params  = var.json_params_content
#   tags         = var.tags

#   /* lifecycle {
#     prevent_destroy = true  //var.is_productive  # : allow destruction only for non production environment
#   } */

#   timeouts {
#     create = var.operation_timeout
#     update = var.operation_timeout
#     delete = var.operation_timeout
#   }
# }

# resource "cloudfoundry_service_key" "service_keys" {
#   for_each         = var.service_keys
#   name             = try( each.value.append_release_to_name, false ) ? "${each.value}_${var.deployed_version}" : each.value
#   service_instance = cloudfoundry_service_instance.service_instance.id

# /*   triggers = {
#     trigger_condition = try( each.value.retrigger_at_each_release, false ) ? var.deployed_version : ""
#   } */
# }


# output "service_keys" {
#   value = cloudfoundry_service_key.service_keys
#   sensitive = true
# }


# output "id" {
#   value = cloudfoundry_service_instance.service_instance.id
# }
