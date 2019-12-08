---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_service_instance"
sidebar_current: "docs-cf-resource-service-instance"
description: |-
  Provides a Cloud Foundry Service Instance.
---

# cloudfoundry\_service\_instance

Provides a Cloud Foundry resource for managing Cloud Foundry [Service Instances](https://docs.cloudfoundry.org/devguide/services/) within spaces.

## Example Usage

The following is a Service Instance created in the referenced space with the specified service plan.

```
data "cloudfoundry_service" "redis" {
    name = "p-redis"
}

resource "cloudfoundry_service_instance" "redis1" {
  name = "pricing-grid"
  space = "${cloudfoundry_space.dev.id}"
  service_plan = "${data.cloudfoundry_service.redis.service_plans["shared-vm"]}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The name of the Service Instance in Cloud Foundry
* `service_plan` - (Required, String) The ID of the [service plan](/docs/providers/cloudfoundry/d/service.html)
* `space` - (Required, String) The ID of the [space](/docs/providers/cloudfoundry/r/space.html)
* `json_params` - (Optional, String) Json string of arbitrary parameters. Some services support providing additional configuration parameters within the provision request. By default, no params are provided.
* `tags` - (Optional, List) List of instance tags. Some services provide a list of tags that Cloud Foundry delivers in [VCAP_SERVICES Env variables](https://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html#VCAP-SERVICES). By default, no tags are assigned.
* `recursive_delete` - (Optional, Bool) Default: `false`. If set `true`, Cloud Foundry will delete any service bindings, service keys, and route mappings associated with the service instance. This flag should only be set when such dependent resources were provisioned outside of terraform, and need removal to enable deletion of the associated service instance.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service instance

## Import

An existing Service Instance can be imported using its guid, e.g.

```
$ terraform import cloudfoundry_service.redis a-guid
```

## Timeouts

* `create` - Default: 15 mins. Terraform will return an error if the resource was not deployed in the given timeframe.
* `delete` - Default: 15 mins. Terraform will return an error if the resource was not deleted in the given timeframe.
* `update` - Default: 15 mins. Terraform will return an error if the resource was not dupdated in the given timeframe.
