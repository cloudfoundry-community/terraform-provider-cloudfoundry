---
layout: "cf"
page_title: "Cloud Foundry: cf_org"
sidebar_current: "docs-cf-resource-org"
description: |-
  Provides a Cloud Foundry Org resource.
---

# cf\_org

Provides a Cloud Foundry resource for managing Cloud Foundry [organizations](https://docs.cloudfoundry.org/concepts/roles.html), assigning quota definitions, and members. 

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.


## Example Usage

The following example creates an org with a specific org-wide quota.

```
resource "cf_org" "o1" {
    name = "organization-one"
    quota = "${cf_quota.runaway.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Org in Cloud Foundry
* `quota` - (Optional) The ID of quota or plan to be given to this Org. By default, no quota is assigned to the org.  
* `managers` - (Optional) List of users to assign [OrgManager](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to. By default, no managers are assigned.
* `billing_managers` - (Optional) List of ID of users to assign [BillingManager](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to.  By default, no billing managers are assigned.
* `auditors` - (Optional) List of ID of users to assign [OrgAuditor](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to.  By default, no auditors are assigned.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the organization
* `quota` - If a quota is not referenced as an argument then the default quota GUID will be exported 

## Import

An existing Organization can be imported using its guid, e.g.

```
$ terraform import cf_org.o1 a-guid
```
