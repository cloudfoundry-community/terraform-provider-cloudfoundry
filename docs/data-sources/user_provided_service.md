---
layout: "cloudfoundry"
page_title: "Cloud Foundry: user_provided_service"
sidebar_current: "docs-cf-datasource-user-provided-service"
description: |-
  Get information on a Cloud Foundry User Provided Service.
---

# cloudfoundry\_user\_provided\_service

Gets information on a Cloud Foundry user provided service (see https://docs.cloudfoundry.org/devguide/services/user-provided.html).

## Example Usage

```hcl
data "cloudfoundry_user_provided_service" "my-instance" {
    name  = "my-service-instance"
    space = "space-id"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user provided service (NOT its guid).
* `space` - (Required) The space GUID in which the user provided service is defined.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service instance.
* `name` - The name of the service instance.
* `credentials` - A map of fields that was defined as input to the user provided service via the `-p` option in cf cli.
* `route_service_url` - The url of the route service that should proxy requests to an app (see https://docs.cloudfoundry.org/devguide/services/route-binding.html).
* `syslog_drain_url` - The url of the syslog service to which app logs should be streamed.
* `tags` - Tags set during service instance creations.
