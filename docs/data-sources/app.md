---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_app"
sidebar_current: "docs-cf-datasource-app"
description: |-
  Get information on a Cloud Foundry Application.
---

# cloudfoundry\_app

Gets information on a Cloud Foundry application.

## Example Usage

```hcl
data "cloudfoundry_app" "my-app" {
    name_or_id = "my-app"
    space      = "space-id"
}
```

## Argument Reference

The following arguments are supported:

* `name_or_id` - (Required) The name of the application or its guid.
* `space` - (Required) The space guid of the app.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the application
* `name` - The name of the application.
* `instances` - The number of app instances that you want to start. Defaults to 1.
* `memory` - The memory limit for each application instance in megabytes.
* `disk_quota` - The disk space to be allocated for each application instance in megabytes.
* `stack` - The GUID of the stack the application will be deployed to.
* `buildpack` - The buildpack used to stage the application.
* `command` - The custom start command for the application.
* `enable_ssh` - Whether to enable or disable SSH access to the container.
* `state` - Current state of the app (`stopped` or `running` or `started`).
* `environment` - Key/value pairs of custom environment variables to set in your app.
* `health_check_http_endpoint` - The endpoint for the http health check type.
* `health_check_type` - The health check type which can be one of "`port`", "`process`", "`http`" or "`none`".
* `health_check_timeout` - The timeout in seconds for the health check.
* `labels` - Labels as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).
* `annotations` - Annotations as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).
