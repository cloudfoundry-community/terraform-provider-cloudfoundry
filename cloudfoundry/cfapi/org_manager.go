package cfapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// OrgManager -
type OrgManager struct {
	log *Logger

	config    coreconfig.Reader
	ccGateway net.Gateway

	apiEndpoint string

	repo organizations.OrganizationRepository
}

// CCOrg -
type CCOrg struct {
	ID string

	Name      string `json:"name"`
	Status    string `json:"status,omitempty"`
	QuotaGUID string `json:"quota_definition_guid,omitempty"`
}

// CCOrgResource -
type CCOrgResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCOrg              `json:"entity"`
}

// CCOrgResourceList -
type CCOrgResourceList struct {
	Resources []CCOrgResource `json:"resources"`
}

// OrgRole -
type OrgRole string

// OrgRoleMember -
const OrgRoleMember = OrgRole("users")

// OrgRoleManager -
const OrgRoleManager = OrgRole("managers")

// OrgRoleBillingManager -
const OrgRoleBillingManager = OrgRole("billing_managers")

// OrgRoleAuditor -
const OrgRoleAuditor = OrgRole("auditors")

// NewOrgManager -
func NewOrgManager(config coreconfig.Reader, ccGateway net.Gateway, logger *Logger) (dm *OrgManager, err error) {

	dm = &OrgManager{
		log:         logger,
		config:      config,
		ccGateway:   ccGateway,
		apiEndpoint: config.APIEndpoint(),
		repo:        organizations.NewCloudControllerOrganizationRepository(config, ccGateway),
	}

	if len(dm.apiEndpoint) == 0 {
		return nil, errors.New("API endpoint missing from config file")
	}

	return dm, nil
}

// FindOrg -
func (om *OrgManager) FindOrg(name string) (org CCOrg, err error) {
	orgModel, err := om.repo.FindByName(name)
	if err != nil {
		return CCOrg{}, err
	}

	org.ID = orgModel.GUID
	org.Name = orgModel.Name
	return org, nil
}

// ReadOrg -
func (om *OrgManager) ReadOrg(orgID string) (org CCOrg, err error) {

	resource := &CCOrgResource{}
	path := fmt.Sprintf("%s/v2/organizations/%s", om.apiEndpoint, orgID)
	if err = om.ccGateway.GetResource(path, &resource); err != nil {
		return CCOrg{}, err
	}

	org = resource.Entity
	org.ID = resource.Metadata.GUID
	return org, nil
}

// CreateOrg -
func (om *OrgManager) CreateOrg(name string, quotaID string) (org CCOrg, err error) {
	payload := map[string]interface{}{"name": name}
	if len(quotaID) > 0 {
		payload["quota_definition_guid"] = quotaID
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return CCOrg{}, err
	}

	resource := CCOrgResource{}
	err = om.ccGateway.CreateResource(om.apiEndpoint, "/v2/organizations", bytes.NewReader(body), &resource)
	if err != nil {
		return CCOrg{}, err
	}
	org = resource.Entity
	org.ID = resource.Metadata.GUID
	return org, nil
}

// UpdateOrg -
func (om *OrgManager) UpdateOrg(org CCOrg) (err error) {

	body, err := json.Marshal(org)
	if err != nil {
		return
	}

	path := fmt.Sprintf("%s/v2/organizations/%s", om.apiEndpoint, org.ID)
	request, err := om.ccGateway.NewRequest("PUT", path, om.config.AccessToken(), bytes.NewReader(body))
	if err != nil {
		return err
	}

	resource := &CCOrgResource{}
	_, err = om.ccGateway.PerformRequestForJSONResponse(request, resource)
	return err
}

// AddUser -
func (om *OrgManager) AddUser(orgID string, userID string, role OrgRole) (err error) {
	path := fmt.Sprintf("/v2/organizations/%s/%s/%s", orgID, role, userID)
	return om.ccGateway.UpdateResource(om.apiEndpoint, path, strings.NewReader(""))
}

// RemoveUser -
func (om *OrgManager) RemoveUser(orgID string, userID string, role OrgRole) (err error) {
	path := fmt.Sprintf("/v2/organizations/%s/%s/%s", orgID, role, userID)
	err = om.ccGateway.DeleteResource(om.apiEndpoint, path)
	if err != nil {
		if strings.HasSuffix(err.Error(), "Please delete the user associations for your spaces in the org.") {
			om.log.DebugMessage("removing user '%s' from all spaces associated with org '%s'", userID, role, orgID)

			spaceRepo := spaces.NewCloudControllerSpaceRepository(om.config, om.ccGateway)
			err = spaceRepo.ListSpacesFromOrg(orgID, func(space models.Space) bool {

				om.log.DebugMessage("Deleting user '%s' from space '%s'", userID, space.GUID)
				err = om.ccGateway.DeleteResource(om.apiEndpoint,
					fmt.Sprintf("/v2/users/%s/spaces/%s", userID, space.GUID))
				if err != nil {
					om.log.DebugMessage("WARNING! removing user '%s' from space '%s': %s", userID, space.GUID, err.Error())
				}
				return true
			})
			if err == nil {

				err = om.ccGateway.DeleteResource(om.apiEndpoint,
					fmt.Sprintf("/v2/organizations/%s/%s/%s", orgID, role, userID))

				if err != nil {
					om.log.DebugMessage("WARNING: removing user '%s' having role '%s' from org '%s' failed: %s",
						userID, role, orgID, err.Error())
				}
			}
			err = nil
		}
	}
	return err
}

// ListUsers -
func (om *OrgManager) ListUsers(orgID string, role OrgRole) (userIDs []interface{}, err error) {
	userList := &CCUserList{}
	path := fmt.Sprintf("%s/v2/organizations/%s/%s", om.apiEndpoint, orgID, role)
	if err = om.ccGateway.GetResource(path, userList); err != nil {
		return userIDs, err
	}
	for _, r := range userList.Resources {
		userIDs = append(userIDs, r.Metadata.GUID)
	}
	return userIDs, nil
}

// DeleteOrg -
func (om *OrgManager) DeleteOrg(id string) (err error) {
	path := fmt.Sprintf("/v2/organizations/%s", id)
	return om.ccGateway.DeleteResource(om.apiEndpoint, path)
}
