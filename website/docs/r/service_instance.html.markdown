---
layout: "cf"
page_title: "Cloud Foundry: cf_service_instance"
sidebar_current: "docs-cf-resource-service-instance"
description: |-
  Provides a Cloud Foundry Service Instance.
---

# cf\_service\_instance

Provides a Cloud Foundry resource for managing Cloud Foundry [Service Instances](https://docs.cloudfoundry.org/devguide/services/) within spaces.

## Example Usage

The following is a Service Instance created within the referenced space and service plan. 

```
data "cf_service" "redis" {
    name = "p-redis"
}

resource "cf_service_instance" "redis1" {
  name = "pricing-grid"
  space = "${cf_space.dev.id}"
  service_plan = "${data.cf_service.redis.service_plans["shared-vm"]}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The name of the Service Instance in Cloud Foundry
* `service_plan` - (Required, String) The ID of the [service plan](/docs/providers/cloudfoundry/d/service_plan.html)
* `space` - (Required, String) The ID of the [space](/docs/providers/cloudfoundry/r/space.html) 
* `json_params` - (Optional, String) Json string of arbitrary parameters. Some services support providing additional configuration parameters within the provision request
* `tags` - (Optional, List) List of instance tags. Some services provide a list of tags that Cloud Foundry delivers in [VCAP_SERVICES Env variables](https://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html#VCAP-SERVICES)
* `recursive_delete` - (Optional, Bool) Default: `false`. If set `true`, Cloud Foundry will delete service bindings, service keys, and routes associated with the service instance too.
## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service instance

## Import

The current Service Instance can be imported using the `service_broker`, e.g.

```
$ terraform import cf_service.redis a-guid
```
