---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_org_quota"
sidebar_current: "docs-cf-datasource-org-quota"
description: |-
  Get information on a Cloud Foundry org Quota.
---

# cloudfoundry\_org\_quota

Gets information on a Cloud Foundry org quota.

## Example Usage

The following example looks up a quota named 'myquota'
identified by its name.

```
data "cloudfoundry_org_quota" "q" {
    name = "myquota"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the quota to look up

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the quota
