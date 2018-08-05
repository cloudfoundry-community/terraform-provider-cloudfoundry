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


## Example Usage

The following example create an isolation segment named `public_exposure`

```
resource "cloudfoundry_isolation_segment" "public" {
  name = "public_exposure"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The name of the segment as declared in `cf` deployment.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the segment

## Import

An existing segment can be imported using its guid, e.g.

```
$ terraform import cloudfoundry_isolation_segment.s1 a-guid
```
