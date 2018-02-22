---
layout: "cf"
page_title: "Cloud Foundry: cf_private_domain_access"
sidebar_current: "docs-cf-resource-private-domain-access"
description: |-
  Provides a Cloud Foundry private domain acceess resource.
---

# cf\_private\_domain\_access

Provides a resource for managing CLoud Foundry [organization](https://docs.cloudfoundry.org/concepts/roles.html)
access to [private domains](https://docs.cloudfoundry.org/devguide/deploy-apps/routes-domains.html#domains).

## Example Usage

The following is an example of giving an organization access to a private [domain]. The
domain is retrieved via a [domain data source](/docs/providers/cloudfoundry/d/domain.html)
and the organization via a [org data source)(/docs/providers/cloudfoundry/d/org.html).

```
resource "cf_private_domain_access" "shared-to-my-org" {
  domain = "${data.cf_domain.domain.id}"
  org    = "${data.cf_org.my-org.id}"
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required, String) The GUID of private domain.
* `org`    - (Required, String) The GUID of the organization.

## Import

The `cf_private_domain_access` can be imported using `<org-guid>/<domain-guid>' identifier, e.g.

```
$ terraform import cf_private_domain_access.my-access 84f5ba83-1728-481f-9a62-72d109e4be74/c8eba5e6-5a21-45ee-ae0a-59b1f650888a
```
