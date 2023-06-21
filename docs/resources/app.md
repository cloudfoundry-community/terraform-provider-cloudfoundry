---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_app"
sidebar_current: "docs-cf-resource-app"
description: |-
  Provides a Cloud Foundry Application resource.
---

# cloudfoundry\_app

Provides a Cloud Foundry [application](https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html) resource.

## Example Usage

The following example creates an application.

```hcl
resource "cloudfoundry_app" "spring-music" {
    name = "spring-music"
    path = "/Work/cloudfoundry/apps/spring-music/build/libs/spring-music.war"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the application.
* `space` - (Required) The GUID of the associated Cloud Foundry space.
* `instances` - (Optional, Number) The number of app instances that you want to start. Defaults to 1.
* `memory` - (Optional, Number) The memory limit for each application instance in megabytes. If not provided, value is computed and retreived from Cloud Foundry.
* `disk_quota` - (Optional, Number) The disk space to be allocated for each application instance in megabytes. If not provided, default disk quota is retrieved from Cloud Foundry and assigned.
* `stack` - (Optional) The name of the stack the application will be deployed to. Use the [`cloudfoundry_stack`](website/docs/d/stack.html.markdown) data resource to lookup the available stack names to override Cloud Foundry default.
* `buildpack` - (Optional, String) The buildpack used to stage the application. There are multiple options to choose from:
  * a Git URL (e.g. [https://github.com/cloudfoundry/java-buildpack.git](https://github.com/cloudfoundry/java-buildpack.git)) or a Git URL with a branch or tag (e.g. [https://github.com/cloudfoundry/java-buildpack.git#v3.3.0](https://github.com/cloudfoundry/java-buildpack.git#v3.3.0) for v3.3.0 tag)
  * an installed admin buildpack name (e.g. my-buildpack)
  * an empty blank string to use built-in buildpacks (i.e. autodetection)
* `buildpacks` - (Optional, List) Multiple `buildpacks` used to stage the application. When both `buildpack` and `buildpacks` are set, `buildpacks` wins. There are multiple options to choose from:
  * a Git URL (e.g. [https://github.com/cloudfoundry/java-buildpack.git](https://github.com/cloudfoundry/java-buildpack.git)) or a Git URL with a branch or tag (e.g. [https://github.com/cloudfoundry/java-buildpack.git#v3.3.0](https://github.com/cloudfoundry/java-buildpack.git#v3.3.0) for v3.3.0 tag)
  * an installed admin buildpack name (e.g. my-buildpack)
* `command` - (Optional, String) A custom start command for the application. This overrides the start command provided by the buildpack.
* `enable_ssh` - (Optional, Boolean) Whether to enable or disable SSH access to the container. Default is `true` unless disabled globally.
* `timeout` - (Optional, Number) Max wait time for app instance startup, in seconds. Defaults to 60 seconds.
* `stopped` - (Optional, Boolean) Defines the desired application state. Set to `true` to have the application remain in a stopped state. Default is `false`, i.e. application will be started.
* `labels` - (Optional, map string of string) Add labels as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).
  Works only on cloud foundry with api >= v3.63.
* `annotations` - (Optional, map string of string) Add annotations as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).
  Works only on cloud foundry with api >= v3.63.

### Application Source / Binary

One of the following arguments must be declared to locate application source or archive to be pushed.

* `path` - (Required) An uri or path to target a zip file. this can be in the form of unix path (`/my/path.zip`) or url path (`http://zip.com/my.zip`)
* `source_code_hash` - (Optional) Used to trigger updates. Must be set to a base64-encoded SHA256 hash of the path specified. The usual way to set this is `${base64sha256(file("file.zip"))}`,
  where "file.zip" is the local filename of the lambda function source archive.

* `docker_image` - (Optional, String) The URL to the docker image with tag e.g registry.example.com:5000/user/repository/tag or docker image name from the public repo e.g. redis:4.0
* `docker_credentials` - (Optional) Defines login credentials for private docker repositories
  * `username` - (Required, String) Username for the private docker repo
  * `password` - (Required, String) Password for the private docker repo

~> **NOTE:** [terraform-provider-zipper](https://github.com/ArthurHlt/terraform-provider-zipper)
can create zip file from `tar.gz`, `tar.bz2`, `folder location`, `git repo` locally or remotely and provide `source_code_hash`.

Example Usage with zipper

```hcl
provider "zipper" {
  skip_ssl_validation = false
}

resource "zipper_file" "fixture" {
  source = "https://github.com/orange-cloudfoundry/gobis-server.git#v1.7.3"
  output_path = "path/to/gobis-server.zip"
}

resource "cloudfoundry_app" "gobis-server" {
    name = "gobis-server"
    path = zipper_file.fixture.output_path
    source_code_hash = zipper_file.fixture.output_sha
    buildpack = "go_buildpack"
}
```

### Application Deployment strategy

* `strategy` - (Required) Strategy to use for creating/updating application. Defaults to `none`
Supported options:
  * `none`:
    * Alias: `standard`, `v2`
    * Description: perform restage/create/restart **with** interruption
  * `blue-green`:
    * Alias: `blue-green-v2`
    * Description: perform restage and create app **without** interruption and rollback if an error occurred (using the "venerable" blue-green pattern commonly used with CAPI v2, requires double the overall app memory available in quota)
  * `rolling`:
    * Description: perform restage and create app **without** interruption and rollback if an error occurred (using the `rolling` strategy provided in CAPI v3, requires memory for a single app instance available in quota)

### Service bindings

* `service_binding` - (Optional, Array) Service instances to bind to the application.
  * `service_instance` - (Required, String) The service instance GUID.
  * `params` - (Optional, Map) A list of key/value parameters used by the service broker to create the binding. Defaults to empty map.

~> **NOTE:** Modifying this argument will cause the application to be restaged.
~> **NOTE:** Resource only manages service binding previously set by resource.

### Routing

* `routes` - (Optional, block) The routes to map to the application to control its ingress traffic.
  * `route` - (Required, String) The route id. Route can be defined using the `cloudfoundry_route` resource
  * `port` - (Number) The port of the application to map the tcp route to.

~> **NOTE:** in the future, the `route` block will support the `port` attribute illustrated above to allow mapping of tcp routes, and listening on custom or multiple ports.  
~> **NOTE:** Route mappings can be controlled from either the `cloudfoundry_app.routes` or the `cloudfoundry_routes.target` attributes. Using both syntaxes will cause conflicts and result in unpredictable behavior.  
~> **NOTE:** A given route can not currently be mapped to more than one application using the `cloudfoundry_app.routes` syntax. As an alternative, use the `cloudfoundry_route.target` syntax instead in this specific use-case.  
~> **NOTE:** Resource only manages route mapping previously set by resource.

#### Example usage

```hcl
resource "cloudfoundry_app" "java-spring" {
# [...]
 routes {
    route = cloudfoundry_route.java-spring.id
 }
 routes {
    route = cloudfoundry_route.java-spring-2.id
 }
}
```

### Environment Variables

* `environment` - (Optional, Map) Key/value pairs of custom environment variables to set in your app. Does not include any [system or service variables](http://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html#app-system-env).

~> **NOTE:** Modifying this argument will cause the application to be restaged.

### Health Checks

* `health_check_http_endpoint` -(Optional, String) The endpoint for the http health check type. The default is '/'.
* `health_check_type` - (Optional, String) The health check type which can be one of "`port`", "`process`", "`http`". Default is "`port`".
* `health_check_timeout` - (Optional, Number) The timeout in seconds for the health check.
* `health_check_invocation_timeout` - (Optional, Number) The timeout in seconds for the health check when app is running. The default is 1s.

## Attributes Reference

The following attributes are exported along with any defaults for the inputs attributes.

* `id` - The GUID of the application
* `id_bg` - The GUID of the application updated by resource when strategy is blue-green.

This allows changes to a resource linked to app resource id to be updated when app will be recreated.

## Timeouts

* App instance startup timeout - see the `timeout` argument.
* App staging timeout - 15 mins.
* Service binding timeout - 5 mins.

## Import

The current App can be imported using the `app` GUID, e.g.

```bash
terraform import cloudfoundry_app.spring-music a-guid
```

## Update resource using blue-green app id

This is an example of usage of `id_bg` attribute to update your resource on a changing app id by blue-green:

```hcl
resource "cloudfoundry_app" "test-app-bg" {
    space            = "apaceid"
    buildpack        = "abuildpack"
    name             = "test-app-bg"
    path             = "myapp.zip"
    strategy         = "blue-green-v2"
    routes {
      route = "arouteid"
    }
}

resource "cloudfoundry_app" "test-app-bg2" {
  space            = "apaceid"
  buildpack        = "abuildpack"
  name             = "test-app-bg2"
  path             = "myapp.zip"
  strategy         = "blue-green-v2"
  routes {
    route = "arouteid"
  }
}

resource "cloudfoundry_network_policy" "my-policy" {
  policy {
    destination_app = cloudfoundry_app.test-app-bg2.id_bg
    port            = "8080"
    protocol        = "tcp"
    source_app      = cloudfoundry_app.test-app-bg.id_bg
  }
}

# When you change either test-app-bg or test-app-bg2 this will affect my-policy to be updated because it use `id_bg` instead of id
```
