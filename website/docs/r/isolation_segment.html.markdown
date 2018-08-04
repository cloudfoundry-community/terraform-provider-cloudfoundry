---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_isolation_segment (Experimental)"
sidebar_current: "docs-cf-resource-isolation_segment"
description: |-
  Provides a Cloud Foundry Isolation segment resource (Experimental).
---

# cloudofundry\_isolation_segment

(Experimental) Provides a Cloud Foundry resource for managing Cloud Foundry [isolation segment](http://v3-apidocs.cloudfoundry.org/version/3.53.0/index.html#isolation-segments).

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.

## Example Usage

```
resource "cloudfoundry_isolation_segment" "public" {
  name = "public_exposure"
  orgs = [
    "${data.cloudfloundry_org.o1.id}",
    "${data.cloudfloundry_org.o2.id}"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The name of the segment as declared in `cf` deployment.
* `orgs` - (Optional, List)   List of ID of origanizations that are entitled with this segment. An
           organization must be entitled with the segment in order to bind it to one space.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the segment


## Import

An existing segment can be imported using its guid, e.g.

```
$ terraform import cloudfoundry_isolation_segment.s1 a-guid
```
