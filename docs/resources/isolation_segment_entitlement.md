---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_isolation_segment_entitlement"
sidebar_current: "docs-cf-resource-isolation-segment-entitlement"
description: |-
  Provides a Cloud Foundry relationship resource between organizations and a isolation_segment.
---

# cloudfoundry\_isolation\_segment\_entitlement

Provides a Cloud Foundry resource for managing Cloud Foundry relationships between an
[isolation segment](https://docs.cloudfoundry.org/adminguide/isolation-segments.html)
and an organization.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.

~> **NOTE:** This resource is experimental and subject to breaking changes.

## Example Usage

The following example entitles the segment `public` to organizations `org1` and `org2`

```hcl
resource "cloudfoundry_isolation_segment_entitlement" "public" {
  segment = data.cloudfoundry_isolation_segment.public.id
  orgs = [
    data.cloudfloundry_org.org1.id,
    data.cloudfloundry_org.org2.id
  ]
}
```

## Argument Reference

The following arguments are supported:

* `segment` - (Required, String) The ID of the isolation segment to entitle.
* `orgs`    - (Required, List)   List of ID of origanizations that are entitled with this segment. An
              organization must be entitled with the segment in order to bind it to one space.
* `default` - (Optional, bool) Set this isolation segment defined as default segment for those organizations. Default to false.

## Attributes Reference

The following attributes are exported:

* `id`   - The GUID of the segment
* `orgs` - The list of organization GUIDs entitled with this segment
* `default` - True if isolation segment defined as default segment for those organizations


## Import

Import not yet supported.
