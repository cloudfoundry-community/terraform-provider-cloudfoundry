---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_space_asgs"
sidebar_current: "docs-cf-resource-space-asgs"
description: |-
  Manage application security groups of a space.
---

# cloudfoundry\_space\_asgs

Provides a Cloud Foundry resource for managing Cloud Foundry application security groups for a space.

~> **NOTE:** This resource requires the acting user of Terraform to be authorized as space developer for the target space. If the target space is managed by terraform, use the resource cloudfoundry_space_users to add the acting user as developer first.
~> **NOTE:** This resource only modifies asgs managed in the resource itself. It ignores any existing default asgs (asgs defined at the platform level).

## Example Usage

The following example assigns one asg1 for running and asg2 for staging to space1. All resources like asg1, asg2 and space1 need to be declared and created before.

```hcl
resource "cloudfoundry_space_asgs" "spaceasgs1" {
    space = cloudfoundry_space.space1.id
    running_asgs = [ cloudfoundry_asg.asg1.id ]
    staging_asgs = [ cloudfoundry_asg.asg2.id ]
}
```

## Argument Reference

The following arguments are supported:

* `space` - (Required) The guid of the target space.
* `running_asgs` - (Optional) List of running [application security groups](/docs/providers/cloudfoundry/r/asg.html) to apply to applications running within this space. Defaults to empty list.
* `staging_asgs` - (Optional) List of staging [application security groups](/docs/providers/cloudfoundry/r/asg.html) to apply to applications being staged for this space. Defaults to empty list.
