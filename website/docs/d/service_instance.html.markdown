---
layout: "cf"
page_title: "Cloud Foundry: cf_service_instance"
sidebar_current: "docs-cf-datasource-service-instance"
description: |-
  Get information on a Cloud Foundry Service.
---

# cf\_service\_instance

Gets information on a Cloud Foundry service instance.

## Example Usage

The following example looks up a service instance named 'mydb' within the given space reference. 

```
data "cf_service_instance" "mydb" {
    name = "mydb"
    space = "${data.cf_space.dev.id}
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The name of the service instance to look up.
* `space` - (Optional, String) The space within which the service instance has been created.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service
