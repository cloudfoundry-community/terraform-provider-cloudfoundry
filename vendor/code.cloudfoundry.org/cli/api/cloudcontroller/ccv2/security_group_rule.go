package ccv2

import (
	"code.cloudfoundry.org/cli/types"
)

// SecurityGroupRule represents a Cloud Controller Security Group Role.
type SecurityGroupRule struct {
	// Description is a short message discribing the rule.
	Description string

	// Destination is the destination CIDR or range of IPs.
	Destination string

	// Ports is the port or port range.
	Ports string

	// Protocol can be tcp, icmp, udp, all.
	Protocol string

	// control signal for icmp, where -1 allows all
	Type types.NullInt

	// control signal for icmp, where -1 allows all
	Code types.NullInt

	// enables logging for the egress rule, only valid for tcp rules
	Log types.NullBool
}
