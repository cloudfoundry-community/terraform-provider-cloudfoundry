data "cloudfoundry_stack" "my_stack" {
  name = "cflinuxfs3"
}

data "cloudfoundry_domain" "domain_mydomain" {
  name = "cfapps.eu12.hana.ondemand.com"
}