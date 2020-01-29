---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_route"
sidebar_current: "docs-cf-datasource-route"
description: |-
  Get information on a Cloud Foundry route.
---

# cloudfoundry\_route

Gets information on a Cloud Foundry route.

## Example Usage

```hcl
data "cloudfoundry_route" "my-route" {
    domain   = "domain-id"
    hostname = "my-host"
}
```

## Argument Reference

The following arguments are supported and will be used to perform the lookup:

* `domain` - (Required) The domain guid associated to the route.
* `hostname` - (Optional) The hostname associated to the route to lookup.
* `org` - (Optional) The org guid associated to the route to lookup.
* `path` - (Optional) The path associated to the route to lookup.
* `port` - (Optional) The port associated to the route to lookup.


## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the route.
* `hostname` - (Optional) The hostname associated to the route.
* `path` - (Optional) The path associated to the route.
* `port` - (Optional) The port associated to the route.
