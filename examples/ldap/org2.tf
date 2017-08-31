#
# Configur Org2 with Development space and add users from ldap group 'org2'
#

data "ldap_query" "org2-managers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=managers,ou=org2,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cf_user" "org2-manager" {
  count = "${length(data.ldap_query.org2-managers.results)}"
  name  = "${data.ldap_query.org2-managers.results[count.index]}"

  depends_on = ["cf_user.pcf-users"]
}

data "ldap_query" "org2-developers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=developers,ou=org2,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cf_user" "org2-developer" {
  count = "${length(data.ldap_query.org2-developers.results)}"
  name  = "${data.ldap_query.org2-developers.results[count.index]}"

  depends_on = ["cf_user.pcf-users"]
}

resource "cf_org" "org2" {
  name     = "org2"
  managers = ["${data.cf_user.org2-manager.*.id}"]
}

resource "cf_space" "org2-dev" {
  name       = "dev"
  org        = "${cf_org.org2.id}"
  developers = ["${data.cf_user.org2-developer.*.id}"]
}
