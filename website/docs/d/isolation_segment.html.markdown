---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_isolation_segment"
sidebar_current: "docs-cf-datasource-isolation_segment"
description: |-
  Get information on a Cloud Foundry Isolation segment.
---

# cloudfoundry\_isolation_segment

Gets information on a Cloud Foundry Isolation segment.

## Example Usage

The following example looks up a segment named 'public-exposure'.

```
data "cloudfoundry_isolation_segment" "public" {
    name = "public_exposure"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the segment to look up.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the segment
