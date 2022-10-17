terraform {
  required_version = ">= 1.1"
  required_providers {
    cloudfoundry = {
      source  = "github.wdf.sap.corp/intelligent-rpa/cloudfoundry"
      version = "2.0.0-alpha"
    }
  }
}

variable "sso_passcode" {}

provider "cloudfoundry" {
  # api_url           = "https://api.cf.sap.hana.ondemand.com" 
  api_url           = "https://api.cf.eu12.hana.ondemand.com/"
  sso_passcode      = var.sso_passcode
  store_tokens_path = "tokens.txt"
}

# data "cloudfoundry_app_v3" "app_test" {
#   name_or_id = "approuter"
#   space      = "21352b88-590f-44f9-99c1-ad5967644764"
# }

data "cloudfoundry_org" "org" {
  name = "IPA-CloudOps_ipa-cloudops--infra"
}

data "cloudfoundry_space" "space" {
  name = "main"
  org  = data.cloudfoundry_org.org.id
}

output "org_test_output" {
  value = data.cloudfoundry_org.org
}

output "space_test_v3_output" {
  value = data.cloudfoundry_space.space
}
