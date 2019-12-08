---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_org"
sidebar_current: "docs-cf-datasource-space"
description: |-
  Get information on a Cloud Foundry Space.
---

# cloudfoundry\_space

Gets information on a Cloud Foundry space.

## Example Usage

The following example looks up a space named 'myspace' within an organization 'myorg' which has been previously provisioned thru Terraform.

```
data "cloudfoundry_space" "s" {
    name = "myspace"
    org = "${cloudfoundry_org.myorg.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the space to look up.

One of the following arguments must be provided.

* `org` - (Optional) GUID of the organization the space belongs to. Conflicts with `org_name`.
* `org_name` - (Optional) Name of the organization the space belongs to. Conflicts with `org`.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the space
* `org` - The GUID of the org the space belongs to
* `org_name` - The name of the org the space belongs to
* `quota`- The GUID of the space's quota
* `labels` - Map of labels as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
Works only on cloud foundry with api >= v3.63.
* `annotations` - Map of annotations as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
Works only on cloud foundry with api >= v3.63.
