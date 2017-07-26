---
layout: "cf"
page_title: "Cloud Foundry: cf_users"
sidebar_current: "docs-cf-resource-users"
description: |-
  Provides a Cloud Foundry bulk user import mport resource.
---

# cf\_users

Provides a Cloud Foundry resource for bulk importing users.

## Example Usage

> TO BE IMPLEMENTED

The following example imports users in bulk and validates each using the given LDAP query. Any user that does not validate against the LDAP query will be added as a UAA user.

```
resource "cf_users" "admin-service-user" {
  
  users = [ 
    "roger.moore@acme.com|Roger|Moore",
    "sean.connery@acme.com|Sean|Connery",
    "matt.damon@acme.com|Matt|Damon",
    "johnny.depp@acme.com|Johnny|Depp",
  ]
  
  ldap {
    
    url = "ldap://ldap.acme.com:3268"
    bind_dn = "admin@acme.com"
    bind_password = "****"

    search_base_dn = "dc=acme,dc=com"
    search_query = "(&(objectClass=user)(memberOf=cn=PCF-USERS,ou=WW,ou=Security Groups,ou=x_NewStructure,dc=int,dc=acme,dc=com)(mail={{username}}))"
    
    username_attribute = "mail"
    given_name_attribute = "givenName"
    family_name = "familyName"
  }

  password = "Passw0rd"

  orgs = [ "${cf_org.org1.id}", "${cf_org.org2.id}" ]
}
```

Or import users to a specific organization by querying an LDAP group.

```
resource "cf_users" "admin-service-user" {
  
  ldap {

    url = "ldap://ldap.acme.com:3268"
    bind_dn = "admin@acme.com"
    bind_password = "****"

    search_base_dn = "dc=acme,dc=com"
    search_query = "(&(objectClass=user)(memberOf=cn=PCF-USERS-ORG1,ou=WW,ou=Security Groups,ou=x_NewStructure,dc=int,dc=acme,dc=com))"

    username_attribute = "mail"
    given_name_attribute = "givenName"
    family_name_attribute = "familyName"
  }

  org_pattern = ".*"
}
```

Another use case for this resource would be to bulk import all users in a specific LDAP group but not assign them to any organization via the `orgs` attribute. Instead use the [`cf_user_org_role`](/docs/providers/cf/r/user_org_role.html) resource to assign subsets of users to specific organizations.

## Argument Reference

The following arguments are supported:

* `users` - (Optional) List of users to import in bulk. Optionally their given and family names can be provided using the '|' separator as shown in the example above. If the given and family names are provided in this manner this will override any attributes retrieved via LDAP.

* `password` - (Optional) A common password to create for each imported user.
* `origin` - (Optional) The user authentcation origin. By default this will be `UAA`. 

* `personal_org` - (Optional) Create a personal org for each user with the same name as the username.
  - `quota` - (Required) The ID of the quota to apply to the personal org

### LDAP User Import

* `ldap` - (Optional) An LDAP query to either validate users against or retrieve the list of users for bulk insertion.
  - `url` - (Required) The LDAP endpoint
  - `bind_dn` - (Required) The bind user name
  - `bind_password` - (Required) The bind user's password
  - `search_base_dn` - (Optional)  The base DN for the query
  - `search_query` - (Required) The search query.
  - `username_attribute` - (Required) The username LDAP attribute
  - `given_name_attribute` - If LDAP user then extract the user's given name from this LDAP attribute.
  - `family_name_attribute` - If LDAP user then extract the user's family name from this LDAP attribute.

### Org Association

Only one of the following arguments should be provided.

* `orgs` - (Optional) List of [orgs](org.html) these users will be a associated with.
* `org_pattern` - (Optional) A regex pattern for retrieving all the orgs the users should be associated with.

## Attributes Reference

The following attributes are exported:

* `users` - A map of imported user GUIDs keyed by the username.
