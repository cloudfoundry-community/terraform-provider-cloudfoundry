---
layout: "cf"
page_title: "Cloud Foundry: cf_network_policy"
sidebar_current: "docs-cf-resource-network-policy"
description: |-
  Provides a Cloud Foundry Network Policy resource.
---

# cf\_network\_policy

Provides a Cloud Foundry resource for managing Cloud Foundry network policies to manage
access between applications via [container-to-container networking](https://docs.cloudfoundry.org/devguide/deploy-apps/cf-networking.html).


## Example Usage

The following creates container to container access policy between the given applications.

```
resource "cf_network_policy" "my-policy" {

    policy {
        source_app = "${cf_app.app1.id}"
        destination_app = "${cf_app.app2.id}"
        port = "8080-8090"
    }

    policy {
        source_app = "${cf_app.app1.id}"
        destination_app = "${cf_app.app3.id}"
        port = "9999"
        protocol = "udp"
    }
}
```

## Argument Reference

The following arguments are supported:

- `policy` - (Required, List) List of policies that allow direct network traffic from one app to another.
  
  - `source_app` - (Required, String) The ID of the [application](/docs/providers/cf/r/app.html) to connect from.
  - `destination_app` - (Required, String) The ID of the [application](/docs/providers/cf/r/app.html) to connect to.
  - `port` - (Required, String) Port (8080) or range of ports (8080-8085) for connection to destination app
  - `protocol` - (Optional, String) One of 'udp' or 'tcp' identifying the allowed protocol for the access. Default is 'tcp'.

## Attributes Reference

The following attributes are exported along with any defaults for the inputs attributes.

* `id` - The GUID of the network_policy

## Import

The current Network policy can be imported using the `network_policy`, e.g.

```
$ terraform import cf_network_policy.my-policy a-guid
```
