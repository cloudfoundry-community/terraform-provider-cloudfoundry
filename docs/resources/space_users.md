---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_space_users"
sidebar_current: "docs-cf-resource-space-users"
description: |- Provides a Cloud Foundry Space users resource.
---

# cloudfoundry\_space\_users

Provides a Cloud Foundry resource for managing Cloud Foundry [space](https://docs.cloudfoundry.org/concepts/roles.html)
members.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted at least with `OrgManager`
permission.
~> **NOTE:** Only modify users managed in the resource, and ignore any existing other users provisioned elsewhere if not
using `force` attribute.

## Example Usage

The following example creates an org with a specific org-wide quota.

```hcl
resource "cloudfoundry_space_users" "su1" {
  space      = "space-id"
  managers   = [
    data.cloudfoundry_user.tl.id,
    "username",
  ]
  developers = [
    data.cloudfoundry_user.tl.id,
    data.cloudfoundry_user.dev1.id,
    data.cloudfoundry_user.dev2.id,
    "username",
  ]
  auditors   = [
    data.cloudfoundry_user.adr.id,
    data.cloudfoundry_user.dev3.id,
    "username"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `space` - (Required) Space associated guid.
* `managers` - (Optional) List of users to
  assign [SpaceManager](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to. Defaults to empty list.
* `developers` - (Optional) List of users to
  assign [SpaceDeveloper](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to. Defaults to empty list.
* `auditors` - (Optional) List of users to
  assign [SpaceAuditor](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to. Defaults to empty list.
* `force` - (Optional, Boolean) Set to true to enforce that users defined here will be only theses users defined (remove
  users roles from external modification).

~> **NOTE:** User can be either an uua guid or a username as cloud foundry treat them both as valid identifier

## Import

An existing Users list can be imported using its space guid, e.g.

```bash
terraform import cloudfoundry_space_users.su1 space-guid
```
