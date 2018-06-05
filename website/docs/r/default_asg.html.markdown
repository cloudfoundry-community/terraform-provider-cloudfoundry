---
layout: "cf"
page_title: "Cloud Foundry: cf_default_asg"
sidebar_current: "docs-cf-resource-default-asg"
description: |-
  Provides a Cloud Foundry Default Application Security Group resource.
---

# cf\_default\_asg

Provides a resource for modifying the default staging or running
[application security groups](https://docs.cloudfoundry.org/adminguide/app-sec-groups.html).

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.

## Example Usage

The following example shows how to apply [application security groups](/docs/providers/cloudfoundry/r/asg.html)
defined elsewhere in the Terraform configuration, to the default running set.  

```
resource "cf_default_asg" "running" {
    name = "running"
    asgs = [ "${cf_asg.messaging.id}", "${cf_asg.services.id}" ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) This should be one of `running` or `staging`
* `asgs` - (Required) A list of references to application security groups IDs.

## Import

The current Default Asg can be imported using the `name` (either `running` or `staging` constant) e.g. 

```
$ terraform import cf_default_asg.running <running/staging>
```