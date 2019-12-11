---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_default_asg"
sidebar_current: "docs-cf-resource-default-asg"
description: |-
  Provides a Cloud Foundry Default Application Security Group resource.
---

# cloudfoundry\_default\_asg

Provides a resource for modifying the default staging or running
[application security groups](https://docs.cloudfoundry.org/adminguide/app-sec-groups.html).

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.

## Example Usage

The following example shows how to apply [application security groups](/docs/providers/cloudfoundry/r/asg.html)
defined elsewhere in the Terraform configuration, to the default running set.  

```hcl
resource "cloudfoundry_default_asg" "running" {
    name = "running"
    asgs = [ "${cloudfoundry_asg.messaging.id}", "${cloudfoundry_asg.services.id}" ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) This should be one of `running` or `staging`
* `asgs` - (Required) A list of references to application security groups IDs.

## Import

The current Default Asg can be imported using the `name` (either `running` or `staging` constant) e.g. 

```bash
$ terraform import cloudfoundry_default_asg.running <running/staging>
```