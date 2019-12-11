---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_service_instance"
sidebar_current: "docs-cf-datasource-service-instance"
description: |-
  Get information on a Cloud Foundry Service instance.
---

# cloudfoundry\_service\_instance

Gets information on a Cloud Foundry service instance.

## Example Usage

```
data "cloudfoundry_service_instance" "my-instance" {
    name_or_id = "my-service-instance"
    space      = "space-id"
}
```

## Argument Reference

The following arguments are supported:

* `name_or_id` - (Required) The name of the service instance or its guid.
* `space` - (Required) The space guid of the app.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service instance.
* `name` - The name of the service instance.
* `service_plan_id` - The service plan used by the service instance.
* `tags` - Tags set during service instance creations.
