---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_user"
sidebar_current: "docs-cf-datasource-user"
description: |-
  Get information on a Cloud Foundry User.
---

# cloudfoundry\_user

Gets information on a Cloud Foundry user.

## Example Usage

The following example looks up an user named 'myuser'.

```hcl
data "cloudfoundry_user" "myuser" {
    name = "myuser"    
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user to look up
* `org_id` - (Optional) the org ID where to look for this user. Use this in case you are using an OrgAdmin account.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the user
