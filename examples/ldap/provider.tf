#
# CF Provider
#

provider "cloudfoundry" {
  api_url             = "https://api.local.pcfdev.io"
  user                = "admin"
  password            = "admin"
  uaa_client_id       = "admin"
  uaa_client_secret   = "admin-client-secret"
  skip_ssl_validation = true
}

#
# LDAP Provider
#

provider "ldap" {
  host          = "localhost"
  port          = 40389
  bind_dn       = "cn=admin,dc=example,dc=org"
  bind_password = "admin"
}
