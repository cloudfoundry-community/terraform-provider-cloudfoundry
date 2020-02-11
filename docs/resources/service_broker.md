---
layout: "cloudfoundry"
page_title: "Cloud Foundry: cloudfoundry_service_broker"
sidebar_current: "docs-cf-resource-service-broker"
description: |-
  Provides a Cloud Foundry Service Broker resource.
---

# cloudfoundry\_service\_broker

Provides a Cloud Foundry resource for managing [service brokers](https://docs.cloudfoundry.org/services/) definitions. 

~> **NOTE:** To visibility of service plans provided by a registred service brijer, use the [cloudfoundry_service_plan_access](service_plan_access.html) resource. 
~> **NOTE:** This resource requires the provider to be authenticated with an account granted org manager permissions.
~> **NOTE:** If catalog is accessible to terraform broker will be automatically updated if catalog change from previous version in resource.

## Example Usage

The following example registers a service broker.

```hcl
resource "cloudfoundry_service_broker" "mysql" {
	name = "test-mysql"
	url = "http://mysql-broker.local.pcfdev.io"
	username = "admin"
	password = "admin"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the service broker
* `url` - (Required) The URL to the service broker [API](https://docs.cloudfoundry.org/services/api.html)
* `username` - (Required) The user name to use to authenticate against the service broker API calls
* `password` - (Required) The password to authenticate against the service broker API calls
* `space` - (Optional) The ID of the space to scope this broker to (registering the broker as [space-scoped](http://docs.cloudfoundry.org/services/managing-service-brokers.html#register-broker)). By default, registers [standard](http://docs.cloudfoundry.org/services/managing-service-brokers.html#register-broker) brokers 
* `fail_when_catalog_not_accessible` - (Optional) Set to true if you want to see errors when getting service broker catalog (default behaviour is silently failed).

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the service broker
* `service_plans` - Map of service plan GUIDs keyed by service "&lt;service name&gt;/&lt;plan name&gt;"
* `services` - Map of service service GUIDs keyed by service name

## Import

An existing Service Broker can be imported using its guid, e.g.

```bash
$ terraform import cloudfoundry_service_broker.mysql a-guid
```
