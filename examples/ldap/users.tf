#
# Get all users in pcf group and add them to cloud foundry
# 

data "ldap_query" "pcf-users" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=users,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid", "givenName", "sn", "mail"]
  index_attribute = "uid"
}

resource "cloudfoundry_user" "pcf-users" {
  count = "${length(data.ldap_query.pcf-users.results)}"

  name     = "${data.ldap_query.pcf-users.results[count.index]}"
  password = "Passw0rd"

  given_name  = "${data.ldap_query.pcf-users.results_attr[join("/",list(data.ldap_query.pcf-users.results[count.index],"givenName"))]}"
  family_name = "${data.ldap_query.pcf-users.results_attr[join("/",list(data.ldap_query.pcf-users.results[count.index],"sn"))]}"
  email       = "${data.ldap_query.pcf-users.results_attr[join("/",list(data.ldap_query.pcf-users.results[count.index],"mail"))]}"
}
