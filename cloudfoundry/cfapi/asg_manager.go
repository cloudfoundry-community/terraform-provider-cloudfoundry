package cfapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/api/securitygroups"
	running "code.cloudfoundry.org/cli/cf/api/securitygroups/defaults/running"
	staging "code.cloudfoundry.org/cli/cf/api/securitygroups/defaults/staging"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// ASGManager -
type ASGManager struct {
	log *Logger

	config    coreconfig.Reader
	ccGateway net.Gateway

	apiEndpoint string

	repo        securitygroups.SecurityGroupRepo
	runningRepo running.SecurityGroupsRepo
	stagingRepo staging.SecurityGroupsRepo
}

// CCASGRule -
type CCASGRule struct {
	Protocol    string `json:"protocol"`
	Destination string `json:"destination"`
	Ports       string `json:"ports,omitempty"`
	Code        int    `json:"code,omitempty"`
	Type        int    `json:"type,omitempty"`
	Log         bool   `json:"log,omitempty"`
	Description string `json:"description,omitempty"`
}

// CCASG -
type CCASG struct {
	ID               string
	Name             string      `json:"name"`
	Rules            []CCASGRule `json:"rules"`
	IsRunningDefault bool        `json:"running_default"`
	IsStagingDefault bool        `json:"staging_default"`
}

// CCASGResource -
type CCASGResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCASG              `json:"entity"`
}

// NewASGManager -
func newASGManager(config coreconfig.Reader, ccGateway net.Gateway, logger *Logger) (dm *ASGManager, err error) {
	dm = &ASGManager{
		log:         logger,
		config:      config,
		ccGateway:   ccGateway,
		apiEndpoint: config.APIEndpoint(),
		repo:        securitygroups.NewSecurityGroupRepo(config, ccGateway),
		runningRepo: running.NewSecurityGroupsRepo(config, ccGateway),
		stagingRepo: staging.NewSecurityGroupsRepo(config, ccGateway),
	}

	if len(dm.apiEndpoint) == 0 {
		return nil, errors.New("API endpoint missing from config file")
	}

	return dm, nil
}

// CreateASG -
func (am *ASGManager) CreateASG(name string, rules []CCASGRule) (id string, err error) {

	body, err := json.Marshal(map[string]interface{}{
		"name":  name,
		"rules": rules,
	})
	if err != nil {
		return "", err
	}

	resource := CCASGResource{}
	if err = am.ccGateway.CreateResource(am.apiEndpoint, "/v2/security_groups", bytes.NewReader(body), &resource); err != nil {
		return "", err
	}
	id = resource.Metadata.GUID
	return id, nil
}

// UpdateASG -
func (am *ASGManager) UpdateASG(id string, name string, rules []CCASGRule) (err error) {

	body, err := json.Marshal(map[string]interface{}{
		"name":  name,
		"rules": rules,
	})
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/v2/security_groups/%s", am.apiEndpoint, id)
	request, err := am.ccGateway.NewRequest("PUT", path, am.config.AccessToken(), bytes.NewReader(body))
	if err != nil {
		return err
	}

	resource := &CCASGResource{}
	_, err = am.ccGateway.PerformRequestForJSONResponse(request, resource)
	return err
}

// GetASG -
func (am *ASGManager) GetASG(id string) (asg CCASG, err error) {
	resource := &CCASGResource{}
	path := fmt.Sprintf("%s/v2/security_groups/%s", am.apiEndpoint, id)
	if err = am.ccGateway.GetResource(path, resource); err != nil {
		return CCASG{}, err
	}
	asg = resource.Entity
	asg.ID = resource.Metadata.GUID
	return asg, nil
}

// Delete -
func (am *ASGManager) Delete(id string) (err error) {
	return am.ccGateway.DeleteResource(am.apiEndpoint, fmt.Sprintf("/v2/security_groups/%s", id))
}

// Read -
func (am *ASGManager) Read(name string) (models.SecurityGroup, error) {
	return am.repo.Read(name)
}

// Running -
func (am *ASGManager) Running() (asgs []string, err error) {
	securityGroups, err := am.runningRepo.List()
	if err != nil {
		return []string{}, err
	}
	for _, s := range securityGroups {
		asgs = append(asgs, s.GUID)
	}
	return asgs, nil
}

// BindToRunning -
func (am *ASGManager) BindToRunning(id string) error {
	return am.runningRepo.BindToRunningSet(id)
}

// UnbindFromRunning -
func (am *ASGManager) UnbindFromRunning(id string) error {
	return am.runningRepo.UnbindFromRunningSet(id)
}

// UnbindAllFromRunning -
func (am *ASGManager) UnbindAllFromRunning() (err error) {
	securityGroups, err := am.runningRepo.List()
	if err != nil {
		return err
	}
	for _, s := range securityGroups {
		err = am.runningRepo.UnbindFromRunningSet(s.GUID)
		if err != nil {
			return err
		}
	}
	return nil
}

// Staging -
func (am *ASGManager) Staging() (asgs []string, err error) {
	securityGroups, err := am.stagingRepo.List()
	if err != nil {
		return []string{}, err
	}
	for _, s := range securityGroups {
		asgs = append(asgs, s.GUID)
	}
	return asgs, nil
}

// BindToStaging -
func (am *ASGManager) BindToStaging(id string) error {
	return am.stagingRepo.BindToStagingSet(id)
}

// UnbindFromStaging -
func (am *ASGManager) UnbindFromStaging(id string) error {
	return am.stagingRepo.UnbindFromStagingSet(id)
}

// UnbindAllFromStaging -
func (am *ASGManager) UnbindAllFromStaging() (err error) {
	securityGroups, err := am.stagingRepo.List()
	if err != nil {
		return err
	}
	for _, s := range securityGroups {
		err = am.stagingRepo.UnbindFromStagingSet(s.GUID)
		if err != nil {
			return err
		}
	}
	return nil
}
