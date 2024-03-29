---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_user"
sidebar_current: "docs-cf-resource-user"
description: |-
  Provides a Cloud Foundry User resource.
---

# cloudfoundry_user

Provides a Cloud Foundry resource for registering users. This resource provides extended
functionality to attach additional UAA roles to the user.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions and UAA admin client. See related [uaa documentation](http://docs.cloudfoundry.org/uaa/uaa-user-management.html#creating-users)
~> **NOTE:** Create an existing user will not trigger any errors and will lead to retrieve id of existing user and update it.

## Example Usage

The following example creates a user and attaches additional UAA roles to grant administrator rights to that user.

```hcl
resource "cloudfoundry_user" "admin-service-user" {
    
    name = "cf-admin"
    password = "Passw0rd"
    
    given_name = "John"
    family_name = "Doe"

    groups = [ "cloud_controller.admin", "scim.read", "scim.write" ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user. This will also be the users login name
* `password` - (Optional) The user's password
* `origin` - (Optional) The user authentcation origin. By default this will be `UAA`. For users authenticated by LDAP this should be `ldap`
* `given_name` - (Optional) The given name of the user
* `family_name` - (Optional) The family name of the user
* `email` - (Optional) The email address of the user. When not provided, name is used as email.
* `groups` - (Optional) Any UAA `groups` / `roles` to associated the user with

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the User
* `email` - If not provided this attributed will be assigned the same value as the `name`, assuming that the username is the user's email address

## Import

An existing User can be imported using its guid, e.g.

```bash
terraform import cloudfoundry_user.admin-service-user a-guid
```
