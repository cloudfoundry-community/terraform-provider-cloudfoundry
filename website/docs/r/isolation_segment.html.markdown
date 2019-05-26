---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_isolation_segment"
sidebar_current: "docs-cf-resource-isolation-segment"
description: |-
  Provides a Cloud Foundry Isolation segment resource.
---

# cloudfoundry\_isolation\_segment

(Experimental) Provides a Cloud Foundry resource for managing Cloud Foundry
[isolation segment](http://v3-apidocs.cloudfoundry.org/version/3.53.0/index.html#isolation-segments).

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.

~> **NOTE:** This resource is experimental and subject to breaking changes.

See `cloudfoundry_isolation_segment_entitlement` resource to assign the segment to one-or-more
origanizations.

~> **NOTE:** Note: An isolation segment cannot be deleted if it is entitled to any organization.
   Make sure to request deletion of all `cloudfoundry_isolation_segment_entitlement`
   resources prior to request deletion of the associated `cloudfoundry_isolation_segment`

## Example Usage

The following example creates an isolation segment named `public_exposure`

```
resource "cloudfoundry_isolation_segment" "public" {
  name = "public_exposure"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) Isolation Segment names must be unique across the entire system,
  and case is ignored when checking for uniqueness. The name must match the value specified in
  the placement_tags section of the Diego manifest file. If the names do not match, Cloud Foundry
  fails to place apps in the isolation segment when apps are started or restarted in the space
  assigned to the isolation segment.
 * `labels` - (Optional, map string of string) Add labels as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
 Works only on cloud foundry with api >= v3.63.
 * `annotations` - (Optional, map string of string) Add annotations as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
 Works only on cloud foundry with api >= v3.63.


## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the segment

## Import

An existing segment can be imported using its guid, e.g.

```
$ terraform import cloudfoundry_isolation_segment.s1 a-guid
```
