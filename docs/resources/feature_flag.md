---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_feature_flags"
sidebar_current: "docs-cf-resource-feature-flags"
description: |-
  Provides a Cloud Foundry configuration resource.
---

# cloudfoundry\_feature\_flags

Provides a Cloud Foundry configuration resource for managing Cloud Foundry [feature](https://docs.cloudfoundry.org/adminguide/listing-feature-flags.html) flags.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.

## Example Usage

The following is an example updates Cloud Foundry feature flags. Each of the flags will also be computed from current settings and exported if not changed.

```hcl
resource "cloudfoundry_feature_flags" "config" {

  feature_flags {
    user_org_creation                    = false
    private_domain_creation              = true
    app_bits_upload                      = true
    app_scaling                          = true
    route_creation                       = true
    service_instance_creation            = true
    diego_docker                         = false
    set_roles_by_username                = true
    unset_roles_by_username              = true
    task_creation                        = true
    env_var_visibility                   = true
    space_scoped_private_broker_creation = true
    space_developer_env_var_visibility   = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `feature_flags` - (Required) [Feature flags](https://docs.cloudfoundry.org/adminguide/listing-feature-flags.html) enabling/disabling certain Cloud Foundry features.
  * `user_org_creation` - (Optional) Any user can create an organization. Minimum CC API version: 2.12.
  * `private_domain_creation` - (Optional) An Org Manager can create private domains for that organization. Minimum CC API version: 2.12.
  * `app_bits_upload` - (Optional) Space Developers can upload app bits. Minimum CC API version: 2.12.
  * `app_scaling` - (Optional) Space Developers can perform scaling operations (i.e. change memory, disk, or instances). Minimum CC API version: 2.12.
  * `route_creation` - (Optional) Space Developers can create routes in a space. Minimum CC API version: 2.12.
  * `service_instance_creation` - (Optional) Space Developers can create service instances in a space. Minimum CC API version: 2.12.
  * `diego_docker` - (Optional) Space Developers can push Docker apps. Minimum CC API version 2.33.
  * `set_roles_by_username` - (Optional) Org Managers and Space Managers can add roles by username. Minimum CC API version: 2.37.
  * `unset_roles_by_username` - (Optional) Org Managers and Space Managers can remove roles by username. Minimum CC API version: 2.37.
  * `task_creation` - (Optional) Space Developers can create tasks on their application. This feature is under development. Minimum CC API version: 2.47.
  * `env_var_visibility` - (Optional) All users can view environment variables. Minimum CC API version: 2.58.
  * `space_scoped_private_broker_creation` - (Optional) Space Developers can create space-scoped private service brokers. Minimum CC API version: 2.58.
  * `space_developer_env_var_visibility` - (Optional) Space Developers can view their v2 environment variables. Org Managers and Space Managers can view their v3 environment variables. Minimum CC API version: 2.58.

When not provided, optional fields are filled with their actual value in Cloud Foundry.

## Import

The current Config can be imported using the `config` constant, e.g.

```bash
terraform import cloudfoundry_feature_flags.config config
```
