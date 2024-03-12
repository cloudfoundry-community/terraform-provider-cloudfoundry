---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_service_key"
sidebar_current: "docs-cf-resource-service-key"
description: |-
  Provides a Cloud Foundry Service Key.
---

# cloudfoundry\_service\_key

Provides a Cloud Foundry resource for managing Cloud Foundry [Service Keys](https://docs.cloudfoundry.org/devguide/services/#service-keys).

## Example Usage

The following creates a Service Key for the referenced Service Instance.

```hcl
data "cloudfoundry_service" "redis" {
    name = "p-redis"
}

resource "cloudfoundry_service_instance" "redis1" {
  name = "pricing-grid"
  space = cloudfoundry_space.dev.id
  service_plan = data.cloudfoundry_service.redis.service_plans["shared-vm"]
}

resource "cloudfoundry_service_key" "redis1-key1" {
  name = "pricing-grid-key1"
  service_instance = cloudfoundry_service_instance.redis1.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Service Key in Cloud Foundry.
* `service_instance` - (Required) The ID of the Service Instance the key should be associated with.
* `params` - (Optional, Map) A list of key/value parameters used by the service broker to create the binding for the key. By default, no parameters are provided.
* `params_json` - (Optional, String) Arbitrary parameters in the form of stringified JSON object to pass to the service bind handler.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service instance.
* `credentials` - Credentials for this service key that can be used to bind to the associated service instance.

## Import

An existing Service Key can be imported using its guid , e.g.

```bash
terraform import cloudfoundry_service_key.redis1-key1 a-guid
```

### Timeouts

`cloudfoundry_service_key` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `60 seconds`) Used for Creating Instance.
* `delete` - (Default `60 seconds`) Used for Destroying Instance.
