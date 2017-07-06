---
layout: "cf"
page_title: "Cloud Foundry: cf_stack"
sidebar_current: "docs-cf-datasource-stack"
description: |-
  Get information on a Cloud Foundry stack.
---

# cf\_stack

Gets information on a particular Cloud Foundry [stack](https://docs.cloudfoundry.org/devguide/deploy-apps/stacks.html).

## Example Usage

The following example looks up a stack named 'my_custom_stack'. 

```
data "cf_stack" "mystack" {
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
