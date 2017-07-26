---
layout: "cf"
page_title: "Cloud Foundry: cf_app"
sidebar_current: "docs-cf-resource-app"
description: |-
  Provides a Cloud Foundry Application resource.
---

# cf\_app

Provides a Cloud Foundry application resource for managing Cloud Foundry [applications](https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html).

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

* `name` - (Required) The name of the application in Cloud Foundry space.
* `space` - (Required) The GUID of the associated space.
* `ports` - (Optional, Array of Number) A list of ports which the app will listen on.
* `instances` - (Optional, Number) The number of app instances that you want to start.
* `memory` - (Optional, Number) The memory limit for each application instance in megabytes.
* `disk_quota` - (Optional, Number) The disk space to be allocated for each application instance in megabytes.
* `stack` - (Optional) The GUID of the stack the application will be deployed to. Use the [`cf_stack`](/docs/providers/cf/d/stack.html) data resource to lookup the stack GUID to overriding the default.
* `buildpack` - (Optional, String) The custom buildpack to use. This will bypass the buildpack detect phase. There are three options to choose from:
a) Blank means autodetection; b) A Git Url pointing to a buildpack; c) Name of an installed admin buildpack.
* `command` - (Optional, String) A custom start command for the application. This overrides the start command provided by the buildpack.
* `enable_ssh` - (Optional, Boolean) Whether to enable or disable SSH access to the container. Default is `true` unless disabled globally.
* `timeout` - (Optional, Number) Defines the number of seconds that Cloud Foundry waits for starting your application.
* `stopped` - (Optional, Boolean) The application will be created and remain in an stopped state. Default is to stage and start the application.

### Application Source / Binary

One of the following arguments must be declared to locate application source or archive to be pushed.

* `url` - (Optional, String) The URL for the application binary. A local path may be referenced via "`file://...`".

* `git` - (Optional, String) The git location to pull the application source directly from source control.

  - `url` - (Required, String) The git URL for the application repository.
  - `branch` - (Optional, String) The branch of from which the repository contents should be retrieved.
  - `tag` - (Optional, String) The version tag of the contents to retrieve.
  - `key` - (Optional, String) The git private key to access a private repo via SSH.
  - `user` - (Optional, String) Git user for accessing a private repo.
  - `password` - (Optional, String) Git password for accessing a private repo.

      > Arguments "`tag`" and "`branch`" are mutually exclusive. If a git SSH "`key`" is provided and it is protected the "`password`" argument should be used as the key's password.

* `github_release` - (Optional, String) The Buildpack archive published as a github release.

  - `owner` - (Required, String) The github owner or organization name
  - `repo` - (Required, String) The repository containing the release
  - `token` - (Optional, String) Github API token to use to access Github
  - `version` - (Optional, String) The version or tag of the release.
  - `filename` - (Required, String) The name of the published file. The values `zipball` or `tarball` will download the published
  
### Service bindings

Modifying this argument will cause the application to be restaged.

* `service_binding` - (Optional, Array) Service instances to bind to.

  - `service` - (Required, String) The service instance GUID.
  - `params` - (Optional, Map) A list of key/value parameters used by the service broker to create the binding.

### Routing and Blue-Green Deployment Strategy

* `route` - (Optional) Configures how the application or service will be accessed. This can be also used to define a blue-green app deployment strategy. When doing a blue-green deployment the actual name of the application will be timestamped to differentiate between the current live application and the more recent staged application.

  - `default_route` - (Optional, String) The GUID of the default route where the application will be available once deployed.

  > TO BE IMPLEMENTED

  - `stage_route` - (Optional, String) The GUID of the route where the staged application will be available.
  - `live_route` - (Optional, String) The GUID of the route where the live application will be available.
  - `validation_script` - (Optional, String) The validation script to execute against the stage application before mapping the live route to the staged application.

### Environment Variables

Modifying this argument will cause the application to be restaged.

* `environment` - (Optional, Map) Key/value pairs of all the environment variables to run in your app. Does not include any system or service variables.

### Health Checks

* `health_check_http_endpoint` -(Optional, String) The endpoint for the http health check type. The default is '/'.
* `health_check_type` - (Optional, String) The health check type which can be one of "`port`", "`process`", "`http`" or "`none`". Default is "`http`".
* `health_check_timeout` - (Optional, Number) The timeout in seconds for the health check.

## Attributes Reference

The following attributes are exported along with any defaults for the inputs attributes.

* `id` - The GUID of the application

