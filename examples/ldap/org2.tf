#
# Configur Org2 with Development space and add users from ldap group 'org2'
#

data "ldap_query" "org2-managers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=managers,ou=org2,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cloudfoundry_user" "org2-manager" {
  count = "${length(data.ldap_query.org2-managers.results)}"
  name  = "${data.ldap_query.org2-managers.results[count.index]}"

  # Ensure all PCF users have been created before
  # referencing them for addition to org / space
  depends_on = ["cloudfoundry_user.pcf-users"]
}

data "ldap_query" "org2-developers" {
  base_dn = "dc=example,dc=org"
  filter  = "(&(objectClass=inetOrgPerson)(memberOf=cn=developers,ou=org2,ou=pcf,dc=example,dc=org))"

  attributes      = ["uid"]
  index_attribute = "uid"
}

data "cloudfoundry_user" "org2-developer" {
  count = "${length(data.ldap_query.org2-developers.results)}"
  name  = "${data.ldap_query.org2-developers.results[count.index]}"

  # Ensure all PCF users have been created before
  # referencing them for addition to org / space
  depends_on = ["cloudfoundry_user.pcf-users"]
}

resource "cloudfoundry_org" "org2" {
  name     = "org2"
  managers = ["${data.cloudfoundry_user.org2-manager.*.id}"]
}

resource "cloudfoundry_space" "org2-dev" {
  name       = "dev"
  org        = "${cloudfoundry_org.org2.id}"
  developers = ["${data.cloudfoundry_user.org2-developer.*.id}"]
}
