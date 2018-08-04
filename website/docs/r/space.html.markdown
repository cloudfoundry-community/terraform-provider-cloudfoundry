---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_space"
sidebar_current: "docs-cf-resource-space"
description: |-
  Provides a Cloud Foundry Space resource.
---

# cf\_space

Provides a Cloud Foundry resource for managing Cloud Foundry [spaces](https://docs.cloudfoundry.org/concepts/roles.html) within organizations.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted org manager permissions.


## Example Usage

The following is a Space created within the referenced Org. All referenced users must have been added as a member to the owning Org.

```
resource "cloudfoundry_space" "s1" {
    name = "space-one"
    org = "${cloudfoundry_org.o1.id}"
    quota = "${cloudfoundry_quota.dev.id}"
    asgs = [ "${cloudfoundry_asg.svc.id}" ]
    managers = [
        "${cloudfoundry_user_org_role.tl.user}"
    ]
    developers = [
        "${cloudfoundry_user_org_role.tl.user}",
        "${cloudfoundry_user_org_role.dev1.user}",
        "${cloudfoundry_user_org_role.dev2.user}"
    ]
    auditors = [
        "${cloudfoundry_user_org_role.adr.user}",
        "${cloudfoundry_user_org_role.dev3.user}"
    ]
    allow_ssh = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Space in Cloud Foundry.
* `org` - (Required) The ID of the [Org](/docs/providers/cloudfoundry/r/org.html) within which to create the space.
* `quota` - (Optional) The ID of the Space [quota](/docs/providers/cloudfoundry/r/quota.html) or plan defined for the owning Org. By default, no space quota is assigned to the space.
* `allow_ssh` - (Optional) Allows SSH to application containers via the [CF CLI](https://github.com/cloudfoundry/cli)
* `isolation_segment` - (`Experimental`,Optional) The ID of the isolation segment on which the space is bound. The segment must be entitled to the space's parent origanization.
* `asgs` - (Optional) List of running [application security groups](/docs/providers/cloudfoundry/r/asg.html) to apply to applications running within this space.
* `staging_asgs` - (Optional) List of staging [application security groups](/docs/providers/cloudfoundry/r/asg.html) to apply to applications being staged for this space.
* `managers` - (Optional) List of users to assign [SpaceManager](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to.
* `developers` - (Optional) List of users to assign [SpaceDeveloper](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to.
* `auditors` - (Optional) List of users to assign [SpaceAuditor](https://docs.cloudfoundry.org/concepts/roles.html#roles) role to.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the Space

## Import

An existing Space can be imported using its guid, e.g.

```
$ terraform import cloudfoundry_space.s1 a-guid
```
