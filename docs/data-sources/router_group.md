---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_router_group"
sidebar_current: "docs-cf-datasource-router-group"
description: |-
  Get information on a Cloud Foundry router_group.
---

# cloudfoundry\_router\_group

Gets information on a particular Cloud Foundry router group. Router groups are used to declare [TCP domains](https://docs.cloudfoundry.org/devguide/deploy-apps/router_groups.html) and need to be referenced when declaring [TCP routes](https://docs.cloudfoundry.org/adminguide/enabling-tcp-routing.html).

~> **NOTE:** This data source requires the provider to be authenticated with an account granted admin permissions.

## Example Usage

The following example looks up a router group named 'my_custom_router_group'.

```hcl
data "cloudfoundry_router_group" "default-tcp" {
    name = "default-tcp"    
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the router group to look up

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the router group
* `type` - The type of the router group
