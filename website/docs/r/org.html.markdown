---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_org"
sidebar_current: "docs-cf-resource-org"
description: |-
  Provides a Cloud Foundry Org resource.
---

# cloudfoundry\_org

Provides a Cloud Foundry resource for managing Cloud Foundry [organizations](https://docs.cloudfoundry.org/concepts/roles.html), assigning quota definitions, and members. 

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.
~> **NOTE:** Only modify users managed in the resource, and ignore any existing other users provisioned elsewhere

## Example Usage

The following example creates an org with a specific org-wide quota.

```
resource "cloudfoundry_org" "o1" {
    name = "organization-one"
    quota = "${cloudfoundry_quota.runaway.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Org in Cloud Foundry
* `quota` - (Optional) The ID of quota or plan to be given to this Org. By default, no quota is assigned to the org.  
* `labels` - (Optional, map string of string) Add labels as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
Works only on cloud foundry with api >= v3.63.
* `annotations` - (Optional, map string of string) Add annotations as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
Works only on cloud foundry with api >= v3.63.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the organization
* `quota` - If a quota is not referenced as an argument then the default quota GUID will be exported 

## Import

An existing Organization can be imported using its guid, e.g.

```
$ terraform import cloudfoundry_org.o1 a-guid
```
