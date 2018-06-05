---
layout: "cf"
page_title: "Cloud Foundry: cf_private_domain_access"
sidebar_current: "docs-cf-resource-private-domain-access"
description: |-
  Provides a Cloud Foundry private domain access resource.
---

# cf\_private\_domain\_access

Provides a resource for sharing access to [private domains](https://docs.cloudfoundry.org/devguide/deploy-apps/routes-domains.html#domains) with other Cloud Foundry Organizations.

~> **NOTE:** Multiple instances of this resource can be used to share a given private domain with multiple orgs.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted org manager permissions.


## Example Usage

The following is an example of giving an organization access to a private [domain]. The
domain is retrieved via a [domain data source](/docs/providers/cloudfoundry/d/domain.html)
and the organization via an [org data source)(/docs/providers/cloudfoundry/d/org.html).

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

An existing CF private domain access can be imported using the combined `<org-guid>/<domain-guid>' identifier, e.g.

```
$ terraform import cf_private_domain_access.my-access 84f5ba83-1728-481f-9a62-72d109e4be74/c8eba5e6-5a21-45ee-ae0a-59b1f650888a
```
