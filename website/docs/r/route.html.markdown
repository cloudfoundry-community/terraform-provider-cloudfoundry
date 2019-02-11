---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_route"
sidebar_current: "docs-cf-resource-route"
description: |-
  Provides a Cloud Foundry route resource.
---

# cloudfoundry\_route

Provides a Cloud Foundry resource for managing Cloud Foundry application [routes](https://docs.cloudfoundry.org/devguide/deploy-apps/routes-domains.html).

## Example Usage

The following example creates an route for an application.

```
resource "cloudfoundry_route" "default" {
    domain = "${data.cloudfoundry_domain.apps.domain.id}"
    space = "${data.cloudfoundry_space.dev.id}"
    hostname = "myapp"
}
```

## Argument Reference

The following arguments are supported:

- `domain` - (Required, String) The ID of the domain to map the host name to. If not provided the default application domain will be used.
- `space` - (Required, String) The ID of the space to create the route in.
- `hostname` - (Required, Optional) The application's host name. This is required for shared domains.

The following arguments apply only to TCP routes.

- `port` - (Optional, Int) The port to associate with the route for a TCP route. Conflicts with `path` and `random_port`.
- `random_port` - (Optional, Bool) Set to 'true' to create a random port. Conflicts with `path` and `port` and defaults to false.

The following argument applies only to HTTP routes.

- `path` - (Optional) A path for a HTTP route.

The following maps the route to an application.

- `target` - (Optional, Set) One or more route mapping(s) that will map this route to application(s). Can be repeated multiple times to load balance route traffic among multiple applications.<br/>
The `target` block supports:
  - `app` - (Required, String) The ID of the [application](/docs/providers/cloudfoundry/r/app.html) to map this route to.
  - `port` - (Optional, Int) A port that the application will be listening on. If this argument is not provided then the route will be associated with the application's default port. 

~> **NOTE:** Route mappings can be controlled from either the `cloudfoundry_routes.target` or the `cloudfoundry_app.routes` attributes. Using both syntaxes will cause conflicts and result in unpredictable behavior.  


## Attributes Reference

The following attributes are exported along with any defaults for the inputs attributes.

* `id` - The GUID of the route
* `endpoint` - The complete endpoint with path if set for the route

## Import

The current Route can be imported using the `route`, e.g.

```
$ terraform import cloudfoundry_route.default a-guid
```
