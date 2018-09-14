#
# Configur Org1 with Development space and add users from ldap group 'org1'
#

data "ldap_query" "org1-managers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=managers,ou=org1,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cloudfoundry_user" "org1-manager" {
  count = "${length(data.ldap_query.org1-managers.results)}"
  name  = "${data.ldap_query.org1-managers.results[count.index]}"

  # Ensure all PCF users have been created before
  # referencing them for addition to org / space
  depends_on = ["cloudfoundry_user.pcf-users"]
}

data "ldap_query" "org1-developers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=developers,ou=org1,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cloudfoundry_user" "org1-developer" {
  count = "${length(data.ldap_query.org1-developers.results)}"
  name  = "${data.ldap_query.org1-developers.results[count.index]}"

  # Ensure all PCF users have been created before
  # referencing them for addition to org / space
  depends_on = ["cloudfoundry_user.pcf-users"]
}

resource "cloudfoundry_org" "org1" {
  name     = "org1"
  managers = ["${data.cloudfoundry_user.org1-manager.*.id}"]
}

resource "cloudfoundry_space" "org1-dev" {
  name       = "dev"
  org        = "${cloudfoundry_org.org1.id}"
  developers = ["${data.cloudfoundry_user.org1-developer.*.id}"]
}
