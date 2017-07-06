---
layout: "cf"
page_title: "Cloud Foundry: cf_buildpack"
sidebar_current: "docs-cf-resource-buildpack"
description: |-
  Provides a Cloud Foundry Buildpack resource.
---

# cf\_buildpack

Provides a Cloud Foundry resource for managing Cloud Foundry [buildpacks](https://docs.cloudfoundry.org/adminguide/buildpacks.html).

## Example Usage

The following example creates a Cloud Foundry Buildpack .

```
resource "cf_buildpack" "tomee" {
    name = "tomcat-enterprise-edition"
    path = "https://github.com/cloudfoundry-community/tomee-buildpack"
    position = "12"
    enable = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Buildpack.
* `position` - (Optional, Number) Specifies where to place the buildpack in the detection priority list. For more information, see the [Buildpack Detection](https://docs.cloudfoundry.org/buildpacks/detection.html) topic.
* `enabled` - (Optional, Boolean) Specifies whether to allow apps to be pushed with the buildpack, and defaults to true.
* `locked` - (Optional, Boolean) Specifies whether buildpack is locked to prevent further updates, and defaults to false.

### Buildpack location

One of the following arguments must be declared to locate buildpack source or archive to be uploaded.

* `url` - (Optional, String) Specifies the location of the buildpack to upload from. It can be a URL to a zip file, a github repository or a local directory via "`file://...`".

* `git` - (Optional, String) The git location to pull the builpack source directly from source control.

  - `url` - (Required, String) The git URL for the application repository.
  - `branch` - (Optional, String) The branch of from which the repository contents should be retrieved.
  - `tag` - (Optional, String) The version tag of the contents to retrieve.
  - `user` - (Optional, String) Git user for accessing a private repo.
  - `password` - (Optional, String) Git password for accessing a private repo.
  - `key` - (Optional, String) The git private key to access a private repo via SSH.
  
      > Arguments "`tag`" and "`branch`" are mutually exclusive. If a git SSH "`key`" is provided and it is protected the "`password`" argument should be used as the key's password.

* `github_release` - (Optional, String) The Buildpack archive published as a github release.
  - `owner` - (Required, String) The github owner or organization name.
  - `repo` - (Required, String) The repository containing the release.
  - `token` - (Optional, String) Github API token to use to access Github.
  - `version` - (Optional, String) The version or tag of the release.
  - `filename` - (Required, String) The name of the published file. The values `zipball` or `tarball` will download the published  source archive.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the organization
