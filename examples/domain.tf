data "cloudfoundry_stack_v3" "my_stack" {
  name = "cflinuxfs3"
}

data "cloudfoundry_domain_v3" "domain_mydomain" {
  name = "cfapps.eu12.hana.ondemand.com"
}