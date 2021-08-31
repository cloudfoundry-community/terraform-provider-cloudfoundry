---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_buildpack"
sidebar_current: "docs-cf-resource-buildpack"
description: |-
  Provides a Cloud Foundry Buildpack resource.
---

# cloudfoundry\_buildpack

Provides a Cloud Foundry resource for managing Cloud Foundry admin [buildpacks](https://docs.cloudfoundry.org/adminguide/buildpacks.html).

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.

## Example Usage

The following example creates a Cloud Foundry buildpack .

```hcl
resource "cloudfoundry_buildpack" "tomee" {
    name = "tomcat-enterprise-edition"
    path = "https://github.com/cloudfoundry-community/tomee-buildpack/releases/download/v3.17/tomee-buildpack-v3.17.zip"
    position = "12"
    enable = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Buildpack.
* `position` - (Optional, Number) Specifies where to place the buildpack in the detection priority list. For more information, see the [Buildpack Detection](https://docs.cloudfoundry.org/buildpacks/detection.html) topic. When not provided, cloudfoundry assigns a default buildpack position.
* `enabled` - (Optional, Boolean) Specifies whether to allow apps to be pushed with the buildpack, and defaults to true.
* `locked` - (Optional, Boolean) Specifies whether buildpack is locked to prevent further updates, and defaults to false.
* `labels` - (Optional, map string of string) Add labels as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).
  Works only on cloud foundry with api >= v3.63.
* `annotations` - (Optional, map string of string) Add annotations as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).
  Works only on cloud foundry with api >= v3.63.

### Buildpack location

* `path` - (Required) An uri or path to target a zip file. this can be in the form of unix path (`/my/path.zip`) or url path (`http://zip.com/my.zip`)
* `source_code_hash` - (Optional) Used to trigger updates. Must be set to a base64-encoded SHA256 hash of the path specified. The usual way to set this is `base64sha256(file("file.zip"))`,
where "file.zip" is the local filename of the lambda function source archive.

~> **NOTE:** [terraform-provider-zipper](https://github.com/ArthurHlt/terraform-provider-zipper)
can create zip file from `tar.gz`, `tar.bz2`, `folder location`, `git repo` locally or remotely and provide `source_code_hash`.

Example Usage with zipper:

```hcl
provider "zipper" {
  skip_ssl_validation = false
}

resource "zipper_file" "fixture" {
  source = "https://github.com/cloudfoundry-community/tomee-buildpack.git#v3.17"
  output_path = "path/to/tomee-buildpack_v3.17.zip"
}

resource "cloudfoundry_buildpack" "tomee" {
    name = "tomcat-enterprise-edition"
    path = zipper_file.fixture.output_path
    source_code_hash = zipper_file.fixture.output_sha
    position = "12"
    enable = true
}
```

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the buildpack

## Import

The current buildpack can be imported using the `buildpack` guid, e.g.

```bash
terraform import cloudfoundry_buildpack.tomee a-guid
```
