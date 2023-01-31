package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"code.cloudfoundry.org/cli/types"
	"encoding/json"
)

// Quota is the generic definition of a quota.
type Quota struct {
	// GUID is the unique OrganizationQuota identifier.
	GUID string

	// Name is the name of the OrganizationQuota.
	Name string

	//  If an organization can have services that are not free
	NonBasicServicesAllowed bool

	// How many services an organization can have. (-1 represents an unlimited amount)
	TotalServices int

	// How many service keys an organization can have. (-1 represents an unlimited amount)
	TotalServiceKeys types.NullInt

	// How many routes an organization can have. (-1 represents an unlimited amount)
	TotalRoutes int

	// How many routes an organization can have that use a reserved port.
	// These routes count toward total_routes. (-1 represents an unlimited amount)
	TotalReservedRoutePorts types.NullInt

	// How many private domains an organization can have. (-1 represents an unlimited amount)
	TotalPrivateDomains types.NullInt

	// How much memory in megabyte an organization can have.
	MemoryLimit types.NullByteSizeInMb

	// The maximum amount of memory in megabyte an application instance can have. (-1 represents an unlimited amount)
	InstanceMemoryLimit types.NullByteSizeInMb

	// How many app instances an organization can create. (-1 represents an unlimited amount)
	AppInstanceLimit types.NullInt

	// The number of tasks that can be run per app. (-1 represents an unlimited amount)
	AppTaskLimit types.NullInt

	// The owning organization of the space quota
	OrganizationGUID string
}

func (q *Quota) UnmarshalJSON(data []byte) error {
	var ccQ struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name                    string `json:"name"`
			NonBasicServicesAllowed bool   `json:"non_basic_services_allowed"`
			TotalServices           int    `json:"total_services"`
			TotalServiceKeys        *int   `json:"total_service_keys,omitempty"`
			TotalRoutes             int    `json:"total_routes"`
			TotalReservedRoutePorts *int   `json:"total_reserved_route_ports,omitempty"`
			TotalPrivateDomains     *int   `json:"total_private_domains,omitempty"`
			MemoryLimit             int64  `json:"memory_limit"`
			InstanceMemoryLimit     int64  `json:"instance_memory_limit"`
			AppInstanceLimit        *int   `json:"app_instance_limit"`
			AppTaskLimit            *int   `json:"app_task_limit"`
			OrganizationGUID        string `json:"organization_guid,omitempty"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccQ)
	if err != nil {
		return err
	}

	q.GUID = ccQ.Metadata.GUID
	q.Name = ccQ.Entity.Name
	q.NonBasicServicesAllowed = ccQ.Entity.NonBasicServicesAllowed
	q.TotalServices = ccQ.Entity.TotalServices
	q.TotalServiceKeys.ParseIntValue(ccQ.Entity.TotalServiceKeys)
	q.TotalRoutes = ccQ.Entity.TotalRoutes
	q.TotalReservedRoutePorts.ParseIntValue(ccQ.Entity.TotalReservedRoutePorts)
	q.TotalPrivateDomains.ParseIntValue(ccQ.Entity.TotalPrivateDomains)
	if ccQ.Entity.MemoryLimit > 0 {
		toUint := uint64(ccQ.Entity.MemoryLimit)
		q.MemoryLimit.ParseUint64Value(&toUint)
	}
	if ccQ.Entity.InstanceMemoryLimit > 0 {
		toUint := uint64(ccQ.Entity.InstanceMemoryLimit)
		q.InstanceMemoryLimit.ParseUint64Value(&toUint)
	}
	q.AppInstanceLimit.ParseIntValue(ccQ.Entity.AppInstanceLimit)
	q.AppTaskLimit.ParseIntValue(ccQ.Entity.AppTaskLimit)
	q.OrganizationGUID = ccQ.Entity.OrganizationGUID

	return nil
}

func (q Quota) MarshalJSON() ([]byte, error) {
	ccQ := struct {
		Name                    string `json:"name,omitempty"`
		NonBasicServicesAllowed bool   `json:"non_basic_services_allowed"`
		TotalServices           int    `json:"total_services,omitempty"`
		TotalServiceKeys        *int   `json:"total_service_keys,omitempty"`
		TotalRoutes             int    `json:"total_routes,omitempty"`
		TotalReservedRoutePorts *int   `json:"total_reserved_route_ports,omitempty"`
		TotalPrivateDomains     int    `json:"total_private_domains,omitempty"`
		MemoryLimit             int64  `json:"memory_limit,omitempty"`
		InstanceMemoryLimit     int64  `json:"instance_memory_limit,omitempty"`
		AppInstanceLimit        *int   `json:"app_instance_limit,omitempty"`
		AppTaskLimit            *int   `json:"app_task_limit,omitempty"`
		OrganizationGUID        string `json:"organization_guid,omitempty"`
	}{
		Name:                    q.Name,
		NonBasicServicesAllowed: q.NonBasicServicesAllowed,
		TotalServices:           q.TotalServices,
		TotalRoutes:             q.TotalRoutes,
		OrganizationGUID:        q.OrganizationGUID,
	}

	if !q.MemoryLimit.IsSet {
		ccQ.MemoryLimit = -1
	} else {
		ccQ.MemoryLimit = int64(q.MemoryLimit.Value)
	}

	if !q.InstanceMemoryLimit.IsSet {
		ccQ.InstanceMemoryLimit = -1
	} else {
		ccQ.InstanceMemoryLimit = int64(q.InstanceMemoryLimit.Value)
	}

	if q.TotalServiceKeys.IsSet {
		ccQ.TotalServiceKeys = &q.TotalServiceKeys.Value
	}

	if q.TotalReservedRoutePorts.IsSet {
		ccQ.TotalReservedRoutePorts = &q.TotalReservedRoutePorts.Value
	}

	if !q.TotalPrivateDomains.IsSet {
		ccQ.TotalPrivateDomains = -1
	} else {
		ccQ.TotalPrivateDomains = q.TotalPrivateDomains.Value
	}

	if q.AppInstanceLimit.IsSet {
		ccQ.AppInstanceLimit = &q.AppInstanceLimit.Value
	}

	if q.AppTaskLimit.IsSet {
		ccQ.AppTaskLimit = &q.AppTaskLimit.Value
	}

	return json.Marshal(ccQ)
}

// CreateQuota returns an Quota associated with the
// provided quota and for type either SpaceQuota or OrgQuota.
func (client *Client) CreateQuota(quotaType constant.QuotaType, quota Quota) (Quota, Warnings, error) {
	if quotaType == constant.SpaceQuota {
		return client.CreateSpaceQuotaDefinition(quota)
	}
	return client.CreateOrganizationQuota(quota)
}

// GetQuota returns an Quota associated with the
// provided GUID and for type either SpaceQuota or OrgQuota.
func (client *Client) GetQuota(quotaType constant.QuotaType, guid string) (Quota, Warnings, error) {
	if quotaType == constant.SpaceQuota {
		return client.GetSpaceQuotaDefinition(guid)
	}
	return client.GetOrganizationQuota(guid)
}

// GetQuotas returns an Quota list associated with the
// provided filters and for type either SpaceQuota or OrgQuota.
func (client *Client) GetQuotas(quotaType constant.QuotaType, filters ...Filter) ([]Quota, Warnings, error) {
	if quotaType == constant.SpaceQuota {
		return client.GetSpaceQuotaDefinitions(filters...)
	}
	return client.GetOrganizationQuotas(filters...)
}

// UpdateQuota updates the quota with the given GUID and for type either SpaceQuota or OrgQuota.
func (client *Client) UpdateQuota(quotaType constant.QuotaType, quota Quota) (Quota, Warnings, error) {
	if quotaType == constant.SpaceQuota {
		return client.UpdateSpaceQuotaDefinition(quota)
	}
	return client.UpdateOrganizationQuota(quota)
}

// DeleteQuota delete a quota and for type either SpaceQuota or OrgQuota
func (client *Client) DeleteQuota(quotaType constant.QuotaType, guid string) (Warnings, error) {
	if quotaType == constant.SpaceQuota {
		return client.DeleteSpaceQuotaDefinition(guid)
	}
	return client.DeleteOrganizationQuota(guid)
}
