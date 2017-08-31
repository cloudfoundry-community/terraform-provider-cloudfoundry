#
# Configur Org3 with Development space and add users from ldap group 'org3'
#

data "ldap_query" "org3-managers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=managers,ou=org3,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cf_user" "org3-manager" {
  count = "${length(data.ldap_query.org3-managers.results)}"
  name  = "${data.ldap_query.org3-managers.results[count.index]}"

  depends_on = ["cf_user.pcf-users"]
}

data "ldap_query" "org3-developers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=developers,ou=org3,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cf_user" "org3-developer" {
  count = "${length(data.ldap_query.org3-developers.results)}"
  name  = "${data.ldap_query.org3-developers.results[count.index]}"

  depends_on = ["cf_user.pcf-users"]
}

resource "cf_org" "org3" {
  name     = "org3"
  managers = ["${data.cf_user.org3-manager.*.id}"]
}

resource "cf_space" "org3-dev" {
  name       = "dev"
  org        = "${cf_org.org3.id}"
  developers = ["${data.cf_user.org3-developer.*.id}"]
}
