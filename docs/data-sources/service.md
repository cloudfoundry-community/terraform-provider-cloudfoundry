---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_service"
sidebar_current: "docs-cf-datasource-service"
description: |-
  Get information on a Cloud Foundry Service.
---

# cloudfoundry\_service

Gets information on a Cloud Foundry service definition.

## Example Usage

The following example looks up a service definition named 'p-redis', globally. 

```hcl
data "cloudfoundry_service" "redis" {
    name = "p-redis"    
}
```

The following example looks up a service named 'p-redis', registered as a space-scoped service within the specified Space id

```hcl
data "cloudfoundry_service" "redis" {
    name = "p-redis"  
    space = "${cloudfoundry_space.dev.id}"  
}
```

The following example looks up a service named 'p-redis', registered as a space-scoped service within the specified Space id provided by a specific service broker (its guid is hardcoded here because service brokers are not always visible to non admin users of cloud foundry).

```hcl
data "cloudfoundry_service" "redis" {
    name = "p-redis"  
    space = "${cloudfoundry_space.dev.id}"
    service_broker_guid = "5716f06c-b3a2-4e8a-893f-39870b0c9f42"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the service to look up
* `space` - (Optional) The space within which the service is defined (as a [space-scoped](http://docs.cloudfoundry.org/services/managing-service-brokers.html#register-broker) service)
* `service_broker_guid` - (Optional) The guid of the service broker which offers the service. You can use this to filter two equally named services from different brokers.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service
* `service_plans` - Map of service plan GUIDs keyed by service "&lt;service name&gt;/&lt;plan name&gt;"
* `service_broker_name` - The name of the service broker that offers the service.