---
layout: "cf"
page_title: "Cloud Foundry: cf_user_provided_service"
sidebar_current: "docs-cf-resource-user-provided-service"
description: |-
  Provides a Cloud Foundry User Provided Service.
---

# cf\_user_provided_service

Provides a Cloud Foundry resource for managing Cloud Foundry [User Provided Services](https://docs.cloudfoundry.org/devguide/services/user-provided.html) within spaces.

## Example Usage

The following are User Provided Service created within the referenced space.

```
resource "cf_user_provided_service" "mq" {
  name = "mq-server"
  space = "${cf_space.dev.id}"
  credentials = {
    "url" = "mq://localhost:9000"
    "username" = "admin"
    "password" = "admin"
  }
}

resource "cf_user_provided_service" "mail" {
  name = "mail-server"
  space = "${cf_space.dev.id}"
  credentials_json = <<JSON
  {
    "server" : {
      "host" : "smtp.example.com",
      "port" : 25,
      "tls"  : false
    },
    "auth" : {
      "user"     : "login",
      "password" : "secret"
    }
  }
  JSON
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Service Instance in Cloud Foundry
* `space` - (Required) The ID of the [space](/docs/providers/cloudfoundry/r/space.html)
* `credentials` - (Optional) Arbitrary credentials in the form of key-value pairs and delivered to applications via [VCAP_SERVICES Env variables](https://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html#VCAP-SERVICES). Conflicts with `credentials_json`.
* `credentials_json` - (Optional) Same as `credentials` but in the form of a stringified JSON object. Conflicts with `credentials`.
* `syslog_drain_url` - (Optional) URL to which logs for bound applications will be streamed
* `route_service_url` - (Optional) URL to which requests for bound routes will be forwarded. Scheme for this URL must be https
* `recursive_delete` - (Optional, Bool) Default: `false`. If set `true`, Cloud Foundry will delete service bindings, service keys, and routes associated with the service instance too.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service instance

## Import

The current User Provided Service can be imported using the `user_provided_service`, e.g.

```
$ terraform import cf_user_provided_service.mq-server a-guid
```
