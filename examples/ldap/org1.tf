#
# Configur Org1 with Development space and add users from ldap group 'org1'
#

data "ldap_query" "org1-managers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=managers,ou=org1,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cf_user" "org1-manager" {
  count = "${length(data.ldap_query.org1-managers.results)}"
  name  = "${data.ldap_query.org1-managers.results[count.index]}"

  depends_on = ["cf_user.pcf-users"]
}

data "ldap_query" "org1-developers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=developers,ou=org1,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cf_user" "org1-developer" {
  count = "${length(data.ldap_query.org1-developers.results)}"
  name  = "${data.ldap_query.org1-developers.results[count.index]}"

  depends_on = ["cf_user.pcf-users"]
}

resource "cf_org" "org1" {
  name     = "org1"
  managers = ["${data.cf_user.org1-manager.*.id}"]
}

resource "cf_space" "org1-dev" {
  name       = "dev"
  org        = "${cf_org.org1.id}"
  developers = ["${data.cf_user.org1-developer.*.id}"]
}
