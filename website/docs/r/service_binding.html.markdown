---
layout: "cf"
page_title: "Cloud Foundry: cf_service_binding"
sidebar_current: "docs-cf-resource-service-binding"
description: |-
  Provides a Cloud Foundry resource to bind a service instance.
---

# cf\_service\_binding

Provides a Cloud Foundry resource for delayed [binding](https://docs.cloudfoundry.org/devguide/services/managing-services.html#bind) of service instances to applications.

## Example Usage

The following example enables access to a specific plan of a given service broker within an Org.

```
resource "cf_service_binding" "org1-mysql-512mb" {
    app_instance = "${cf_app.myapp.id}"
    services_instance = "${cf_service_instance.mysql-test.id}"
}
```

## Argument Reference

The following arguments are supported:

* `app_instance` - (Required) The ID of the applicaiton instance to bind the service instance to.
* `services_instance` - (Required) The ID the service instance to bind to the application instance.
