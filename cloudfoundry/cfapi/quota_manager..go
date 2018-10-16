package cfapi

import (
	"bytes"
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/net"
)

// QuotaType - Represents type of quota
type QuotaType int

const (
	// SpaceQuota - Represents quota space
	SpaceQuota QuotaType = 0
	// OrgQuota - Represents org quota type
	OrgQuota QuotaType = 1
)

// QuotaManager -
type QuotaManager struct {
	log         *Logger
	config      coreconfig.Reader
	ccGateway   net.Gateway
	apiEndpoint string
}

// CCQuota -
type CCQuota struct {
	ID                      string
	OrgGUID                 string `json:"organization_guid,omitempty"`
	Name                    string `json:"name"`
	AppInstanceLimit        int    `json:"app_instance_limit"`
	AppTaskLimit            int    `json:"app_task_limit"`
	InstanceMemoryLimit     int64  `json:"instance_memory_limit"`
	MemoryLimit             int64  `json:"memory_limit"`
	NonBasicServicesAllowed bool   `json:"non_basic_services_allowed"`
	TotalServices           int    `json:"total_services"`
	TotalServiceKeys        int    `json:"total_service_keys"`
	TotalRoutes             int    `json:"total_routes"`
	TotalReserveredPorts    int    `json:"total_reserved_route_ports"`
	TotalPrivateDomains     int    `json:"total_private_domains,omitempty"`
}

// CCQuotaResource -
type CCQuotaResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCQuota            `json:"entity"`
}

// NewQuotaManager -
func newQuotaManager(config coreconfig.Reader, ccGateway net.Gateway, logger *Logger) (dm *QuotaManager, err error) {
	dm = &QuotaManager{
		log:         logger,
		config:      config,
		ccGateway:   ccGateway,
		apiEndpoint: config.APIEndpoint(),
	}

	if len(dm.apiEndpoint) == 0 {
		return nil, fmt.Errorf("API endpoint missing from config file")
	}

	return dm, nil
}

// getQuotaPath -
func getQuotaPath(t QuotaType) string {
	if t == SpaceQuota {
		return "/v2/space_quota_definitions"
	}
	return "/v2/quota_definitions"
}

// CreateQuota -
func (qm *QuotaManager) CreateQuota(t QuotaType, quota CCQuota) (id string, err error) {
	body, err := json.Marshal(quota)
	if err != nil {
		return
	}
	path := getQuotaPath(t)
	resource := CCQuotaResource{}
	err = qm.ccGateway.CreateResource(qm.apiEndpoint, path, bytes.NewReader(body), &resource)
	if err != nil {
		return "", err
	}
	id = resource.Metadata.GUID
	return id, nil
}

// UpdateQuota -
func (qm *QuotaManager) UpdateQuota(t QuotaType, quota CCQuota) (err error) {
	body, err := json.Marshal(quota)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s/%s", getQuotaPath(t), quota.ID)
	resource := CCQuotaResource{}
	return qm.ccGateway.UpdateResource(qm.apiEndpoint, path, bytes.NewReader(body), &resource)
}

// ReadQuota -
func (qm *QuotaManager) ReadQuota(t QuotaType, id string) (quota CCQuota, err error) {
	resource := CCQuotaResource{}
	url := fmt.Sprintf("%s%s/%s", qm.apiEndpoint, getQuotaPath(t), id)
	if err = qm.ccGateway.GetResource(url, &resource); err != nil {
		return CCQuota{}, err
	}
	quota = resource.Entity
	quota.ID = resource.Metadata.GUID
	return quota, nil
}

// DeleteQuota -
func (qm *QuotaManager) DeleteQuota(t QuotaType, id string) (err error) {
	path := fmt.Sprintf("%s/%s", getQuotaPath(t), id)
	return qm.ccGateway.DeleteResource(qm.apiEndpoint, path)
}

// FindQuotaByName -
func (qm *QuotaManager) FindQuotaByName(t QuotaType, name string, org *string) (quota CCQuota, err error) {
	found := false
	if qm.ccGateway.ListPaginatedResources(qm.apiEndpoint, getQuotaPath(t), CCQuotaResource{},
		func(resource interface{}) bool {
			quotaResource := resource.(CCQuotaResource)
			if (org != nil) && (quotaResource.Entity.OrgGUID != *org) {
				return true
			}
			if quotaResource.Entity.Name == name {
				found = true
				quota = quotaResource.Entity
				quota.ID = quotaResource.Metadata.GUID
				return false
			}
			return true
		}); err != nil {
		return CCQuota{}, err
	}
	if !found {
		return CCQuota{}, errors.NewModelNotFoundError("Quota", name)
	}
	return quota, nil
}
