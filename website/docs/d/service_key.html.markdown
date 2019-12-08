---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_service_key"
sidebar_current: "docs-cf-datasource-service-key"
description: |-
  Get information on a Cloud Foundry Service key.
---

# cloudfoundry\_service\_key

Gets information on a Cloud Foundry service key.

## Example Usage

```
data "cloudfoundry_service_key" "my-key" {
    name             = "my-service-key"
    service_instance = "service-instance-id"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the service key.
* `service_instance` - (Required) The service instance guid for this service key.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service key.
* `credentials` - Map of credentials found inside service key.
