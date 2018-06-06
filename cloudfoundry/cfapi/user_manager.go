package cfapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// UserManager -
type UserManager struct {
	log *Logger

	config     coreconfig.Reader
	uaaGateway net.Gateway
	ccGateway  net.Gateway

	clientToken string

	groupMap      map[string]string
	defaultGroups map[string]byte

	repo api.UserRepository
}

// UAAUser -
type UAAUser struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
	Origin   string `json:"origin,omitempty"`

	Name   UAAUserName    `json:"name,omitempty"`
	Emails []UAAUserEmail `json:"emails,omitempty"`
	Groups []UAAUserGroup `json:"groups,omitempty"`
}

// UAAUserEmail -
type UAAUserEmail struct {
	Value string `json:"value"`
}

// UAAUserName -
type UAAUserName struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

// UAAUserGroup -
type UAAUserGroup struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Type    string `json:"type"`
}

// UAAGroupResourceList -
type UAAGroupResourceList struct {
	Resources []struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	} `json:"resources"`
}

// CCUser -
type CCUser struct {
	ID string

	UserName         string `json:"username"`
	IsAdmin          bool   `json:"admin,omitempty"`
	IsActive         bool   `json:"active,omitempty"`
	DefaultSpaceGUID bool   `json:"default_space_guid,omitempty"`
}

// CCUserResource -
type CCUserResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCUser             `json:"entity"`
}

// CCUserList -
type CCUserList struct {
	Resources []CCUserResource `json:"resources"`
}

// UserRoleInOrg -
type UserRoleInOrg string

// UserIsOrgManager -
const UserIsOrgManager = UserRoleInOrg("managed_organizations")

// UserIsOrgBillingManager -
const UserIsOrgBillingManager = UserRoleInOrg("billing_managed_organizations")

// UserIsOrgAuditor -
const UserIsOrgAuditor = UserRoleInOrg("audited_organizations")

// UserIsOrgMember -
const UserIsOrgMember = UserRoleInOrg("organizations")

// NewUserManager -
func newUserManager(config coreconfig.Reader, uaaGateway net.Gateway, ccGateway net.Gateway, logger *Logger) (um *UserManager, err error) {
	um = &UserManager{
		log:           logger,
		config:        config,
		uaaGateway:    uaaGateway,
		ccGateway:     ccGateway,
		groupMap:      make(map[string]string),
		defaultGroups: make(map[string]byte),
		repo:          api.NewCloudControllerUserRepository(config, uaaGateway, ccGateway),
	}
	return um, nil
}

func (um *UserManager) loadGroups() (err error) {

	uaaEndpoint := um.config.UaaEndpoint()
	if len(uaaEndpoint) == 0 {
		return errors.New("UAA endpoint missing from config file")
	}

	// Retrieve alls groups
	groupList := &UAAGroupResourceList{}
	err = um.uaaGateway.GetResource(
		fmt.Sprintf("%s/Groups", uaaEndpoint),
		groupList)
	if err != nil {
		return err
	}
	for _, r := range groupList.Resources {
		um.groupMap[r.DisplayName] = r.ID
	}

	// Retrieve default scope/groups for a new user by creating
	// a dummy user and extracting the default scope of that user
	username, err := newUUID()
	if err != nil {
		return err
	}
	userResource := UAAUser{
		Username: username,
		Password: "password",
		Origin:   "uaa",
		Emails:   []UAAUserEmail{{Value: "email@domain.com"}},
	}
	body, err := json.Marshal(userResource)
	if err != nil {
		return err
	}
	user := &UAAUser{}
	err = um.uaaGateway.CreateResource(uaaEndpoint, "/Users", bytes.NewReader(body), user)
	if err != nil {
		return err
	}
	err = um.uaaGateway.DeleteResource(uaaEndpoint, fmt.Sprintf("/Users/%s", user.ID))
	if err != nil {
		return err
	}
	for _, g := range user.Groups {
		um.defaultGroups[g.Display] = 1
	}

	return nil
}

// IsDefaultGroup -
func (um *UserManager) IsDefaultGroup(group string) (ok bool) {
	_, ok = um.defaultGroups[group]
	return ok
}

// GetUser -
func (um *UserManager) GetUser(id string) (user *UAAUser, err error) {
	uaaEndpoint := um.config.UaaEndpoint()
	if len(uaaEndpoint) == 0 {
		return nil, errors.New("UAA endpoint missing from config file")
	}

	user = &UAAUser{}
	path := fmt.Sprintf("%s/Users/%s", uaaEndpoint, id)
	if err = um.uaaGateway.GetResource(path, user); err != nil {
		return nil, err
	}

	return user, nil
}

// CreateUser -
func (um *UserManager) CreateUser(
	username string,
	password string,
	origin string,
	givenName string,
	familyName string,
	email string) (user *UAAUser, err error) {

	uaaEndpoint := um.config.UaaEndpoint()
	if len(uaaEndpoint) == 0 {
		return nil, errors.New("UAA endpoint missing from config file")
	}

	userResource := UAAUser{
		Username: username,
		Password: password,
		Origin:   origin,
		Name: UAAUserName{
			GivenName:  givenName,
			FamilyName: familyName,
		},
	}
	if len(email) > 0 {
		userResource.Emails = append(userResource.Emails, UAAUserEmail{email})
	} else {
		userResource.Emails = append(userResource.Emails, UAAUserEmail{username})
	}

	body, err := json.Marshal(userResource)
	if err != nil {
		return nil, err
	}

	user = &UAAUser{}
	err = um.uaaGateway.CreateResource(uaaEndpoint, "/Users", bytes.NewReader(body), user)
	switch httpErr := err.(type) {
	case nil:
	case errors.HTTPError:
		if httpErr.StatusCode() == http.StatusConflict {
			return nil, errors.NewModelAlreadyExistsError("user", username)
		}
		return nil, err
	default:
		return nil, err
	}

	body, err = json.Marshal(resources.Metadata{GUID: user.ID})
	if err != nil {
		return nil, err
	}

	err = um.ccGateway.CreateResource(um.config.APIEndpoint(), "/v2/users", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser -
func (um *UserManager) UpdateUser(
	id string,
	username string,
	givenName string,
	familyName string,
	email string) (user *UAAUser, err error) {

	uaaEndpoint := um.config.UaaEndpoint()
	if len(uaaEndpoint) == 0 {
		return nil, errors.New("UAA endpoint missing from config file")
	}

	userResource := UAAUser{
		Username: username,
		Name: UAAUserName{
			GivenName:  givenName,
			FamilyName: familyName,
		},
	}
	if len(email) > 0 {
		userResource.Emails = append(userResource.Emails, UAAUserEmail{email})
	} else {
		userResource.Emails = append(userResource.Emails, UAAUserEmail{username})
	}

	body, err := json.Marshal(userResource)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/Users/%s", uaaEndpoint, id)
	request, err := um.uaaGateway.NewRequest("PUT", path, um.config.AccessToken(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.HTTPReq.Header.Set("If-Match", "*")

	user = &UAAUser{}
	_, err = um.uaaGateway.PerformRequestForJSONResponse(request, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword -
func (um *UserManager) ChangePassword(
	id string,
	oldPassword string,
	newPassword string) (err error) {

	uaaEndpoint := um.config.UaaEndpoint()
	if len(uaaEndpoint) == 0 {
		return errors.New("UAA endpoint missing from config file")
	}

	body, err := json.Marshal(map[string]string{
		"oldPassword": oldPassword,
		"password":    newPassword,
	})
	if err != nil {
		return err
	}

	request, err := um.uaaGateway.NewRequest("PUT",
		uaaEndpoint+fmt.Sprintf("/Users/%s/password", id),
		um.config.AccessToken(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.HTTPReq.Header.Set("Authorization", um.clientToken)

	response := make(map[string]interface{})
	_, err = um.uaaGateway.PerformRequestForJSONResponse(request, response)
	return err
}

// UpdateRoles -
func (um *UserManager) UpdateRoles(
	id string, scopesToDelete, scopesToAdd []string, origin string) (err error) {

	uaaEndpoint := um.config.UaaEndpoint()
	if len(uaaEndpoint) == 0 {
		return errors.New("UAA endpoint missing from config file")
	}

	for _, s := range scopesToDelete {
		roleID := um.groupMap[s]
		err = um.uaaGateway.DeleteResource(uaaEndpoint,
			fmt.Sprintf("/Groups/%s/members/%s", roleID, id))
		if err != nil {
			return err
		}
	}

	for _, s := range scopesToAdd {
		roleID, exists := um.groupMap[s]
		if !exists {
			return fmt.Errorf("Group '%s' was not found", s)
		}

		var body []byte
		body, err = json.Marshal(map[string]string{
			"origin": origin,
			"type":   "USER",
			"value":  id,
		})
		if err != nil {
			return err
		}

		response := make(map[string]interface{})
		path := fmt.Sprintf("/Groups/%s/members", roleID)
		err = um.uaaGateway.CreateResource(uaaEndpoint, path, bytes.NewReader(body), &response)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddUserToOrg -
func (um *UserManager) AddUserToOrg(userID string, orgID string) error {

	path := fmt.Sprintf("/v2/users/%s/organizations/%s", userID, orgID)
	return um.ccGateway.CreateResource(um.config.APIEndpoint(), path, bytes.NewReader([]byte{}))
}

// RemoveUserFromOrg -
func (um *UserManager) RemoveUserFromOrg(userID string, orgID string) error {
	path := fmt.Sprintf("/v2/users/%s/organizations/%s", userID, orgID)
	return um.ccGateway.DeleteResource(um.config.APIEndpoint(), path)
}

// ListOrgsForUser -
func (um *UserManager) ListOrgsForUser(userID string, orgRole UserRoleInOrg) (orgIDs []string, err error) {
	orgList := &CCOrgResourceList{}
	path := fmt.Sprintf("%s/v2/users/%s/%s", um.config.APIEndpoint(), userID, orgRole)
	if err = um.ccGateway.GetResource(path, orgList); err != nil {
		return []string{}, err
	}

	orgIDs = []string{}
	for _, o := range orgList.Resources {
		orgIDs = append(orgIDs, o.Metadata.GUID)
	}
	return orgIDs, nil
}

// FindByUsername -
func (um *UserManager) FindByUsername(username string) (models.UserFields, error) {
	return um.repo.FindByUsername(username)
}

// Delete -
func (um *UserManager) Delete(userID string) error {
	return um.repo.Delete(userID)
}
