---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_domain"
sidebar_current: "docs-cf-resource-domain"
description: |-
  Provides a Cloud Foundry Domain resource.
---

# cloudfoundry\_domain

Provides a resource for managing shared or private 
[domains](https://docs.cloudfoundry.org/devguide/deploy-apps/routes-domains.html#domains) in Cloud Foundry.

## Example Usage

The following is an example of a shared domain for a sub-domain of the default application domain 
retrieved via a [domain data source](/docs/providers/cloudfoundry/d/domain.html).

```hcl
resource "cloudfoundry_domain" "shared" {
  sub_domain = "dev"
  domain = "${data.cloudfoundry_domain.apps.domain}"
  internal = false
}
```

The following example creates a private domain owned by the Org referenced by `cloudfoundry_org.pcfdev-org.id`.

```hcl
resource "cloudfoundry_domain" "private" {
  name = "pcfdev-org.io"
  org = "${cloudfoundry_org.pcfdev-org.id}"
}
```

~> **NOTE:** To control sharing of a private domain, use the [cloudfoundry_private_domain](private_domain_access.html) resource. 


## Argument Reference

The following arguments are supported:

* `name` - (Optional, String) Full name of domain. If specified then the `sub_domain` and `domain` attributes will be computed from the `name` 
* `sub_domain` - (Optional, String) Sub-domain part of full domain name. If specified the `domain` argument needs to be provided and the `name` will be computed.
* `domain` - (Optional, String) Domain part of full domain name. If specified the `sub_domain` argument needs to be provided and the `name` will be computed.

The following argument applies only to shared domains.

* `router_group` - (Optional, String) The router group GUID, which can be retrieved via the [`cloudfoundry_router_group`](/docs/providers/cloudfoundry/d/stack.html) data resource. You would need to provide this when creating a shared domain for TCP routes.

The following argument applies only to private domains.

* `org` - (Optional, String) The ID of the Org that owns this domain. If specified, this resource will provision a private domain. By default, the provisioned domain is a public (shared) domain.

* `internal` - (Optional, bool) Flag that sets the domain as an internal domain. Internal domains are used for internal app to app networking only. Defaults to "false". Only works on shared domain.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the domain.

## Import

An existing Domain can be imported using its Domain Guid, e.g.

```bash
$ terraform import cloudfoundry_domain.private a-guid
```
