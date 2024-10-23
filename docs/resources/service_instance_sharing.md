---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_service_instance_sharing"
sidebar_current: "docs-cf-resource-service-instance-sharing"
description: |- Sharing a service instance to another space. 
---

# cloudfoundry\_service\_instance\_sharing

Sharing a service instance to another space [Sharing Service Instances](https://docs.cloudfoundry.org/devguide/services/sharing-instances.html) within spaces.

## Example Usage

The following example shares a specific service instance to the given space.

```hcl
data "cloudfoundry_service_instance" "my-redis" {
  name_or_id = "my-redis"
  space = cloudfoundry_space.dev-1.id
}

resource "cloudfoundry_service_instance_sharing" "share-to-dev-2" {
  service_instance_id = data.cloudfoundry_service_instance.my-redis.id
  space_id        = cloudfoundry_space.dev-2.id
}
```

## Argument Reference

The following arguments are supported:

* `service_instance_id` - (Required, String) The ID of the service instance to share.
* `space_id` - (Required, String) The ID of the space to share the service instance with, the space can be in the same or different org.

## Import

Existing Instance Shared can be imported using the composite `id` formed
with service instance's GUID and space's GUID, seperated by a forward slash '/'.

example: `bb4ea411-service-instance-guid/820b9339-space-guid`
