---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_evg"
sidebar_current: "docs-cf-resource-evg"
description: |-
  Provides a Cloud Foundry Environment Variable Group resource.
---

# cloudfoundry\_evg

Provides a resource for modifying the running or staging [environment variable groups](https://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html#evgroups) in Cloud Foundry.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.
~> **NOTE:** Resource will only modify env var group managed by itself (will not destroy or affect previous env vars set outside of terraform).

## Example Usage

The example below shows how to add environment variables to the running environment variable group.

```hcl
resource "cloudfoundry_evg" "running" { 
  name = "running"

  variables = {
    name1 = "value1"
    name2 = "value2"
    name3 = "value3"
    name4 = "value4"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Either `running` or `staging` to indicate the type of environment variable group to update
* `variables` - (Required) A map of name-value pairs of environment variables

## Import

The current Evg can be imported using the `evg` name (either `running` or `staging` constant) e.g.

```bash
terraform import cloudfoundry_evg.private <running/staging>
```
