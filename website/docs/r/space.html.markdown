---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_space"
sidebar_current: "docs-cf-resource-space"
description: |-
  Provides a Cloud Foundry Space resource.
---

# cloudfoundry\_space

Provides a Cloud Foundry resource for managing Cloud Foundry [spaces](https://docs.cloudfoundry.org/concepts/roles.html) within organizations.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted org manager permissions.
~> **NOTE:** Only modify users managed in the resource, and ignore any existing other users provisioned elsewhere.

## Example Usage

The following is a Space created within the referenced Org. All referenced users must have been added as a member to the owning Org (see [related CF doc](https://docs.cloudfoundry.org/concepts/roles.html#users) for additional context)

```
resource "cloudfoundry_space" "s1" {
    name = "space-one"
    org = "${cloudfoundry_org.o1.id}"
    quota = "${cloudfoundry_quota.dev.id}"
    asgs = [ "${cloudfoundry_asg.svc.id}" ]
    allow_ssh = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Space in Cloud Foundry.
* `org` - (Required) The ID of the [Org](/docs/providers/cloudfoundry/r/org.html) within which to create the space.
* `quota` - (Optional) The ID of the Space [quota](/docs/providers/cloudfoundry/r/space_quota.html) or plan defined for the owning Org. Specifying an empty string requests unassigns any space quota from the space. Defaults to empty string.
* `allow_ssh` - (Optional) Allows SSH to application containers via the [CF CLI](https://github.com/cloudfoundry/cli). Defaults to true.
* `isolation_segment` - (`Experimental`,Optional) The ID of the isolation segment to assign to the space. The segment must be entitled to the space's parent organization. If the isolation segment id is unspecified, then Cloud Foundry assigns the space to the org’s default isolation segment if any. Note that existing apps in the space will not run in a newly assigned isolation segment until they are restarted.
* `asgs` - (Optional) List of running [application security groups](/docs/providers/cloudfoundry/r/asg.html) to apply to applications running within this space. Defaults to empty list.
* `staging_asgs` - (Optional) List of staging [application security groups](/docs/providers/cloudfoundry/r/asg.html) to apply to applications being staged for this space. Defaults to empty list.
* `labels` - (Optional, map string of string) Add labels as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
Works only on cloud foundry with api >= v3.63.
* `annotations` - (Optional, map string of string) Add annotations as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
Works only on cloud foundry with api >= v3.63.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the Space

## Import

An existing Space can be imported using its guid, e.g.

```
$ terraform import cloudfoundry_space.s1 a-guid
```
