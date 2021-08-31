---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_org_users"
sidebar_current: "docs-cf-resource-org-users"
description: |- Provides a Cloud Foundry Org users resource.
---

# cloudfoundry\_org\_users

Provides a Cloud Foundry resource for managing Cloud
Foundry [organizations](https://docs.cloudfoundry.org/concepts/roles.html) members.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.
~> **NOTE:** Only modify users managed in the resource, and ignore any existing other users provisioned elsewhere if not
using `force` attribute.

## Example Usage

The following example creates an org with a specific org-wide quota.

```hcl
resource "cloudfoundry_org_users" "ou1" {
  org              = "organization-id"
  managers         = ["user-guid", "username"]
  billing_managers = ["user-guid", "username"]
  auditors         = ["user-guid", "username"]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Required) Organization associated guid.
* `managers` - (Optional) List of users to assign [OrgManager](https://docs.cloudfoundry.org/concepts/roles.html#roles)
  role to. By default, no managers are assigned.
* `billing_managers` - (Optional) List of ID of users to
  assign [BillingManager](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to. By default, no billing
  managers are assigned.
* `auditors` - (Optional) List of ID of users to
  assign [OrgAuditor](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to. By default, no auditors are
  assigned.
* `force` - (Optional, Boolean) Set to true to enforce that users defined here will be only theses users defined (remove
  users roles from external modification).

~> **NOTE:** User can be either an uua guid or a username as cloud foundry treat them both as valid identifier

## Import

An existing Users list can be imported using its organization guid, e.g.

```bash
terraform import cloudfoundry_org_users.ou1 org-guid
```
