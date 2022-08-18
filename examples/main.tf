variable "sso_passcode" {}

provider "cloudfoundry" {
  api_url           = "https://api.cf.sap.hana.ondemand.com"
  sso_passcode      = var.sso_passcode
  store_tokens_path = "tokens.txt"
}

# data "cloudfoundry_app_v3" "app_test" {
#   name_or_id = "approuter"
#   space      = "21352b88-590f-44f9-99c1-ad5967644764"
# }

data "cloudfoundry_org_v3" "org_test" {
  name = "ipa-deploy--od"
}

data "cloudfoundry_space" "s" {
  name = "master"
  org  = data.cloudfoundry_org_v3.org_test.id
}

data "cloudfoundry_space_v3" "space_test" {
  name = "master"
  org  = data.cloudfoundry_org_v3.org_test.id
}

# output "app_test_output" {
#   value = data.cloudfoundry_app_v3.app_test.id
# }

output "org_test_output" {
  value = data.cloudfoundry_org_v3.org_test
}

output "space_test_output" {
  value = data.cloudfoundry_space.s
}

output "space_test_v3_output" {
  value = data.cloudfoundry_space_v3.space_test
}
