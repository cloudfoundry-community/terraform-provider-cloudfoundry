---
layout: "cf"
page_title: "Cloud Foundry: cf_asg"
sidebar_current: "docs-cf-resource-asg"
description: |-
  Provides a Cloud Foundry Application Security Group resource.
---

# cf\_asg

Provides an [application security group](https://docs.cloudfoundry.org/adminguide/app-sec-groups.html) 
resource for Cloud Foundry. This resource defines egress rules that can be applied to containers that 
stage and run applications.

~> **NOTE:** This resource requires the provider to be authenticated with an account granted admin permissions.


## Example Usage

Basic usage

```
resource "cf_asg" "messaging" {

	name = "rmq-service"
	
    rule {
        protocol = "tcp"
        destination = "192.168.1.100"
        ports = "5671-5672,61613-61614,1883,8883"
        log = true
    }
    rule {
        protocol = "tcp"
        destination = "192.168.1.101"
        ports = "5671-5672,61613-61614,1883,8883"
        log = true
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the application security group.
* `rule` - (Required) A list of egress rules with the following arguments.
  - `protocol` - (Required, String) One of `icmp`, `tcp`, `udp`, or `all`.
  - `destination` - (Required, String) The IP address or CIDR block that can receive traffic.
  - `ports` - (Required, String) A single port, comma-separated ports or range of ports that can receive traffic.
  - `type` - (Optional, Integer) Allowed ICMP [type](https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml#icmp-parameters-types). A value of -1 allows all types. Default is -1.
  - `code` - (Optional, Integer) Allowed ICMP [code](https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml#icmp-parameters-codes). . A value of -1 allows all codes. Default is -1.
  - `log` - (Optional, Boolean) Set to `true` to enable logging. For more information on how to configure system logs to be sent to a syslog drain, review the [ASG logging](http://docs.cloudfoundry.org/concepts/asg.html#logging) documentation.
  - `description` - (Optional, String) Description of the rule.

## Attributes Reference

The following attributes are exported:

* `id` - The GUID of the application security group

## Import

The current Asg can be imported using the `asg` guid, e.g.

```
$ terraform import cf_asg.messaging a-guid
```

