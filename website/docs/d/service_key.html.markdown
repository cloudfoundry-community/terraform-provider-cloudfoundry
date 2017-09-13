---
layout: "cf"
page_title: "Cloud Foundry: cf_service_key"
sidebar_current: "docs-cf-datasource-service-key"
description: |-
  Get information on a Cloud Foundry Service.
---

# cf\_service\_key

Gets information on a Cloud Foundry service key.

## Example Usage

The following example looks up a service key named 'mydb-key' within the given space reference. 

```
data "cf_service_instance" "mydb-key" {
    name = "mydb-key"
    space = "${data.cf_space.dev.id}
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The name of the service key to look up.
* `space` - (Optional, String) The space where the service key was created.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service
* `credentials` - The underlying service credentials to use for binding
