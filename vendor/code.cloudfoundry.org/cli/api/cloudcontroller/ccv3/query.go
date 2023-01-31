package ccv3

import (
	"net/url"
	"strings"
)

// QueryKey is the type of query that is being selected on.
type QueryKey string

const (
	// AppGUIDFilter is a query parameter for listing objects by app GUID.
	AppGUIDFilter QueryKey = "app_guids"
	// GUIDFilter is a query parameter for listing objects by GUID.
	GUIDFilter QueryKey = "guids"
	// NameFilter is a query parameter for listing objects by name.
	NameFilter QueryKey = "names"
	// NoRouteFilter is a query parameter for skipping route creation and unmapping existing routes.
	NoRouteFilter QueryKey = "no_route"
	// OrganizationGUIDFilter is a query parameter for listing objects by Organization GUID.
	OrganizationGUIDFilter QueryKey = "organization_guids"
	// SequenceIDFilter is a query parameter for listing objects by sequence ID.
	SequenceIDFilter QueryKey = "sequence_ids"
	// SpaceGUIDFilter is a query parameter for listing objects by Space GUID.
	SpaceGUIDFilter QueryKey = "space_guids"
	// StackFilter is a query parameter for listing objects by stack name
	StackFilter QueryKey = "stacks"
	// ServiceBrokerGUIDsFilter is a query parameter for getting resources according to the service broker GUID
	ServiceBrokerGUIDsFilter QueryKey = "service_broker_guids"
	// ServiceBrokerNamesFilter is a query parameter when getting plans or offerings according to the Service Brokers that it relates to
	ServiceBrokerNamesFilter QueryKey = "service_broker_names"
	// StatesFilter is a query parameter when getting a package's droplets by state
	StatesFilter QueryKey = "states"
	// HostsFilter is a query param for listing objects by hostname
	HostsFilter QueryKey = "hosts"
	// HostFilter is a query param for getting an object with the given host
	HostFilter QueryKey = "host"
	// PathFilter is a query param for getting an object with the given host
	PathFilter QueryKey = "path"
	// PortFilter is a query param for getting an object with the given port (TCP routes)
	PortFilter QueryKey = "port"

	// FieldsServiceBroker is a query parameter to include specific fields from a service broker in an offering response
	FieldsServiceBroker QueryKey = "fields[service_broker]"
	// FieldsServiceOfferingServiceBroker is a query parameter to include specific fields from a service broker in a plan response
	FieldsServiceOfferingServiceBroker QueryKey = "fields[service_offering.service_broker]"
	// FieldsSpace is a query parameter to include specific fields from a space
	FieldsSpace QueryKey = "fields[space]"
	// FieldsSpaceOrganization is a query parameter to include specific fields from a organization
	FieldsSpaceOrganization QueryKey = "fields[space.organization]"

	// UnmappedFilter is a query parameter specifying unmapped routes
	UnmappedFilter QueryKey = "unmapped"
	// OrderBy is a query parameter to specify how to order objects.
	OrderBy QueryKey = "order_by"
	// PerPage is a query parameter for specifying the number of results per page.
	PerPage QueryKey = "per_page"
	// Include is a query parameter for specifying other resources associated with the
	// resource returned by the endpoint
	Include QueryKey = "include"

	// NameOrder is a query value for ordering by name. This value is used in
	// conjunction with the OrderBy QueryKey.
	NameOrder = "name"

	// PositionOrder is a query value for ordering by position. This value is
	// used in conjunction with the OrderBy QueryKey.
	PositionOrder = "position"

	// Purge is a query parameter used on a Delete request to indicate that dependent resources should also be deleted
	Purge = "purge"

	// SourceGUID is the query parameter for getting an object. Currently it's used as a package GUID
	// to retrieve a package to later copy it to an app (CopyPackage())
	SourceGUID = "source_guid"
)

// Query is additional settings that can be passed to some requests that can
// filter, sort, etc. the results.
type Query struct {
	Key    QueryKey
	Values []string
}

// FormatQueryParameters converts a Query object into a collection that
// cloudcontroller.Request can accept.
func FormatQueryParameters(queries []Query) url.Values {
	params := url.Values{}
	for _, query := range queries {
		params.Add(string(query.Key), strings.Join(query.Values, ","))
	}

	return params
}
