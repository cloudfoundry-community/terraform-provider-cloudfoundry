---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_stack"
sidebar_current: "docs-cf-datasource-stack"
description: |-
  Get information on a Cloud Foundry stack.
---

# cloudfoundry\_stack

Gets information on a particular Cloud Foundry [stack](https://docs.cloudfoundry.org/devguide/deploy-apps/stacks.html).

## Example Usage

The following example looks up a stack named 'my_custom_stack'. 

```
data "cloudfoundry_stack" "mystack" {
    name = "my_custom_stack"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the stack to look up

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the stack
* `description` - The description of the stack
* `labels` - Map of labels as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
Works only on cloud foundry with api >= v3.63.
* `annotations` - Map of annotations as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object). 
Works only on cloud foundry with api >= v3.63.