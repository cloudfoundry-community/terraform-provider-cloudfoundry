---
layout: "cf"
page_title: "Cloud Foundry: cf_service"
sidebar_current: "docs-cf-datasource-service"
description: |-
  Get information on a Cloud Foundry Service.
---

# cf\_service

Gets information on a Cloud Foundry service definition.

## Example Usage

The following example looks up a service definition named 'p-redis', globally. 

```
data "cf_service" "redis" {
    name = "p-redis"    
}
```

The following example looks up a service named 'p-redis', registered as a space-scoped service within the specified Space id

```
data "cf_service" "redis" {
    name = "p-redis"  
    space = "${cf_space.dev.id}"  
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the service to look up
* `space` - (Optional) The space within which the service is defined (as a [space-scoped](http://docs.cloudfoundry.org/services/managing-service-brokers.html#register-broker) service)

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service
* `service_plans` - Map of service plan GUIDs keyed by service "&lt;service name&gt;/&lt;plan name&gt;"