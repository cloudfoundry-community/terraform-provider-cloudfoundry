---
layout: "cf"
page_title: "Cloud Foundry: cf_app"
sidebar_current: "docs-cf-resource-app"
description: |-
  Provides a Cloud Foundry Application resource.
---

# cf\_app

Provides a Cloud Foundry [application](https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html) resource.

## Example Usage

The following example creates an application.

```
resource "cf_app" "spring-music" {
    name = "spring-music"
    url = "file:///Work/cloudfoundry/apps/spring-music/build/libs/spring-music.war"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the application.
* `space` - (Required) The GUID of the associated Cloud Foundry space.
* `instances` - (Optional, Number) The number of app instances to start.
* `memory` - (Optional, Number) The memory limit for each application instance in megabytes.
* `disk_quota` - (Optional, Number) The disk space to be allocated for each application instance in megabytes.
* `stack` - (Optional) The GUID of the stack the application will be deployed to. Use the [`cf_stack`](/docs/providers/cf/d/stack.html) data resource to lookup the stack GUID to overriding the default.
* `buildpack` - (Optional, String) The buildpack used to stage the application. There are multiple options to choose from:
   * a Git URL (e.g. https://github.com/cloudfoundry/java-buildpack.git) or a Git URL with a branch or tag (e.g. https://github.com/cloudfoundry/java-buildpack.git#v3.3.0 for v3.3.0 tag) 
   * an installed admin buildpack name (e.g. my-buildpack)
   * an empty blank string to use built-in buildpacks (i.e. autodetection)
* `command` - (Optional, String) A custom start command for the application. This overrides the start command provided by the buildpack.
* `enable_ssh` - (Optional, Boolean) Whether to enable or disable SSH access to the container. Default is `true`.
* `timeout` - (Optional, Number) Max wait time for app instance startup, in seconds
* `stopped` - (Optional, Boolean) Defines the desired application state. Set to `false` to have the application remain in a stopped state. Default is `false`, i.e. application will be started.

### Application Source / Binary

One of the following arguments must be declared to locate application source or archive to be pushed.

* `url` - (Optional, String) The URL for the application binary. A local path may be referenced via "`file://...`".

* `git` - (Optional, String) The git repository where to pull the application source from.

  - `url` - (Required, String) The git URL for the application repository.
  - `branch` - (Optional, String) The branch of from which the repository contents should be retrieved.
  - `tag` - (Optional, String) The version tag of the contents to retrieve.
  - `key` - (Optional, String) The git private key to access a private repo via SSH.
  - `user` - (Optional, String) Git user for accessing a private repo.
  - `password` - (Optional, String) Git password for accessing a private repo.

~> **NOTE:** Arguments "`tag`" and "`branch`" are mutually exclusive. If a git SSH "`key`" is provided and it is protected the "`password`" argument should be used as the key's password.

* `github_release` - (Optional, String) The github release where to download the application archive from.

  - `owner` - (Required, String) The github owner or organization name
  - `repo` - (Required, String) The repository containing the release
  - `token` - (Optional, String) Github API token to use to access Github
  - `version` - (Optional, String) The version or tag of the release.
  - `filename` - (Required, String) The name of the published file. The values `zipball` or `tarball` will download the published

* `add_content` - (Optional, Array) adds the given content from a local path to the application directory. You can use this attribute to inject files into the pushed application source.

  - `source` - (Required, String) The source path to copy content from. This can be a directory.
  - `destination` - (Required, String) The destination path to copy content to. This is relative to the application source root.

### Service bindings

* `service_binding` - (Optional, Array) Service instances to bind to the application.

  - `service_instance` - (Required, String) The service instance GUID.
  - `params` - (Optional, Map) A list of key/value parameters used by the service broker to create the binding.

~> **NOTE:** Modifying this argument will cause the application to be restaged.   

### Routing

* `route` - (Optional) Configures how the application will be accessed externally to cloudfoundry. 
  - `default_route` - (Optional, String) The ID of the route where the application will be reachable from once deployed.

### Environment Variables

* `environment` - (Optional, Map) Key/value pairs of custom environment variables to set in your app. Does not include any [system or service variables](http://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html#app-system-env). 

~> **NOTE:** Modifying this argument will cause the application to be restaged.

### Health Checks

* `health_check_http_endpoint` -(Optional, String) The endpoint for the http health check type. The default is '/'.
* `health_check_type` - (Optional, String) The health check type which can be one of "`port`", "`process`", "`http`" or "`none`". Default is "`http`".
* `health_check_timeout` - (Optional, Number) The timeout in seconds for the health check.

## Attributes Reference

The following attributes are exported along with any defaults for the inputs attributes.

* `id` - The GUID of the application

## Import

The current App can be imported using the `app` GUID, e.g.

```
$ terraform import cf_app.spring-music a-guid
```

