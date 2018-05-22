---
layout: "cf"
page_title: "Cloud Foundry: cf_service_plan_access"
sidebar_current: "docs-cf-resource-service-access"
description: |-
  Provides a Cloud Foundry Service Access resource.
---

# cf\_service\_access

Provides a Cloud Foundry resource for managing [access](https://docs.cloudfoundry.org/services/access-control.html)
to service plans published by Cloud Foundry [service brokers](https://docs.cloudfoundry.org/services/).

~> **NOTE:** Multiple instances of this resource can be used to share a given service plan with multiple orgs.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.



## Example Usage

The first example enables access to a specific plan of a given service broker to all organizations.
The second example gives access to a specific org.

```
resource "cf_service_plan_access" "org1-mysql-512mb" {
    plan = "${cf_service_broker.mysql.service_plans["p-mysql/512mb"]}"
    public = true
}

resource "cf_service_plan_access" "org1-mysql-512mb" {
    plan = "${cf_service_broker.mysql.service_plans["p-mysql/1gb"]}"
    org = "${cf_org.org1.id}"
}
```

## Argument Reference

The following arguments are supported:

* `plan` - (Required) The ID of the service plan to grant access to
* `org` - (Optional) The ID of the Org which should have access to the plan. Conflicts with `public`.
* `public` - (Optional) Boolean that controls the public state of the plan. Conflicts with `org`.

## Import

The current Service Access can be imported using an `id`.

If given `id` matches existing [`service_plan_visibilities`](https://apidocs.cloudfoundry.org/280/service_plan_visibilities/list_all_service_plan_visibilities.html),
resource will be imported as a `service_plan_access` targeting an organization.

If the given `id` matches [a service plan id](http://apidocs.cloudfoundry.org/280/service_plans/updating_a_service_plan.html),
then the resource will be imported as `service_plan_access` controlling plan's public state.

Otherwise, the import would fail.

E.g.

```
$ terraform import cf_service_plan_access.org1-mysql-512mb a-guid
```
