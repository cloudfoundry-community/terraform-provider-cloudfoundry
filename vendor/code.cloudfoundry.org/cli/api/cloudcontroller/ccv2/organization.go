package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"encoding/json"
	"fmt"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Organization represents a Cloud Controller Organization.
type Organization struct {
	// GUID is the unique Organization identifier.
	GUID string

	// Name is the organization's name.
	Name string

	// QuotaDefinitionGUID is unique identifier of the quota assigned to this
	// organization.
	QuotaDefinitionGUID string

	// DefaultIsolationSegmentGUID is the unique identifier of the isolation
	// segment this organization is tagged with.
	DefaultIsolationSegmentGUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Organization response.
func (org *Organization) UnmarshalJSON(data []byte) error {
	var ccOrg struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name                        string `json:"name"`
			QuotaDefinitionGUID         string `json:"quota_definition_guid,omitempty"`
			DefaultIsolationSegmentGUID string `json:"default_isolation_segment_guid"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccOrg)
	if err != nil {
		return err
	}

	org.GUID = ccOrg.Metadata.GUID
	org.Name = ccOrg.Entity.Name
	org.QuotaDefinitionGUID = ccOrg.Entity.QuotaDefinitionGUID
	org.DefaultIsolationSegmentGUID = ccOrg.Entity.DefaultIsolationSegmentGUID
	return nil
}

type createOrganizationRequestBody struct {
	Name                string `json:"name,omitempty"`
	QuotaDefinitionGUID string `json:"quota_definition_guid,omitempty"`
}

func (client *Client) CreateOrganization(orgName string, quotaGUID string) (Organization, Warnings, error) {
	requestBody := createOrganizationRequestBody{
		Name:                orgName,
		QuotaDefinitionGUID: quotaGUID,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Organization{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostOrganizationRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Organization{}, nil, err
	}

	var org Organization
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &org,
	}

	err = client.connection.Make(request, &response)
	return org, response.Warnings, err
}

// DeleteOrganization deletes the Organization associated with the provided
// GUID. It will return the Cloud Controller job that is assigned to the
// Organization deletion.
func (client *Client) DeleteOrganization(guid string) (Job, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationRequest,
		URIParams:   Params{"organization_guid": guid},
		Query: url.Values{
			"recursive": {"true"},
			"async":     {"true"},
		},
	})
	if err != nil {
		return Job{}, nil, err
	}

	var job Job
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &job,
	}

	err = client.connection.Make(request, &response)
	return job, response.Warnings, err
}

// GetOrganization returns an Organization associated with the provided GUID.
func (client *Client) GetOrganization(guid string) (Organization, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationRequest,
		URIParams:   Params{"organization_guid": guid},
	})
	if err != nil {
		return Organization{}, nil, err
	}

	var org Organization
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &org,
	}

	err = client.connection.Make(request, &response)
	return org, response.Warnings, err
}

// GetOrganizations returns back a list of Organizations based off of the
// provided filters.
func (client *Client) GetOrganizations(filters ...Filter) ([]Organization, Warnings, error) {
	allQueries := ConvertFilterParameters(filters)
	allQueries.Add("order-by", "name")
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationsRequest,
		Query:       allQueries,
	})

	if err != nil {
		return nil, nil, err
	}

	var fullOrgsList []Organization
	warnings, err := client.paginate(request, Organization{}, func(item interface{}) error {
		if org, ok := item.(Organization); ok {
			fullOrgsList = append(fullOrgsList, org)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Organization{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullOrgsList, warnings, err
}

type updateOrgManagerByUsernameRequestBody struct {
	Username string `json:"username"`
}

// UpdateOrganizationManager assigns the org manager role to the UAA user or client with the provided ID.
func (client *Client) UpdateOrganizationManager(guid string, uaaID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationManagerRequest,
		URIParams:   Params{"organization_guid": guid, "manager_guid": uaaID},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// UpdateOrganizationManagerByUsername assigns the org manager role to the user with the provided name.
func (client *Client) UpdateOrganizationManagerByUsername(guid string, username string) (Warnings, error) {
	requestBody := updateOrgManagerByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationManagerByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// UpdateOrganizationBillingManagerByUsername assigns the org manager role to the user with the provided name.
func (client *Client) UpdateOrganizationBillingManagerByUsername(guid string, username string) (Warnings, error) {
	requestBody := updateOrgManagerByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationBillingManagerByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// UpdateOrganizationAuditorByUsername assigns the org manager role to the user with the provided name.
func (client *Client) UpdateOrganizationAuditorByUsername(guid string, username string) (Warnings, error) {
	requestBody := updateOrgManagerByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationAuditorByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// UpdateOrganizationUser makes the user or client with the given UAA ID a
// member of the org.
func (client *Client) UpdateOrganizationUser(guid string, uaaID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationUserRequest,
		URIParams:   Params{"organization_guid": guid, "user_guid": uaaID},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

type updateOrgUserByUsernameRequestBody struct {
	Username string `json:"username"`
}

// UpdateOrganizationUserByUsername makes the user with the given username a member of
// the org.
func (client Client) UpdateOrganizationUserByUsername(orgGUID string, username string) (Warnings, error) {
	requestBody := updateOrgUserByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationUserByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// UpdateOrganization updates the organization with the given GUID.
func (client *Client) UpdateOrganization(orgGuid, orgName, quotaGUID string) (Organization, Warnings, error) {
	requestBody := createOrganizationRequestBody{
		Name:                orgName,
		QuotaDefinitionGUID: quotaGUID,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Organization{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationRequest,
		URIParams:   Params{"organization_guid": orgGuid},
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Organization{}, nil, err
	}

	var updatedObj Organization
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}

// GetOrganizationUsersByRole find all users for an org by role .
// (Only available roles: OrgManager, BillingManager and OrgAuditor)
func (client *Client) GetOrganizationUsersByRole(role constant.UserRole, guid string) ([]User, Warnings, error) {
	requestName := ""
	switch role {
	case constant.OrgManager:
		requestName = internal.GetOrganizationManagersRequest
	case constant.BillingManager:
		requestName = internal.GetOrganizationBillingManagersRequest
	case constant.OrgAuditor:
		requestName = internal.GetOrganizationAuditorsRequest
	case constant.OrgUser:
		requestName = internal.GetOrganizationUsersRequest
	default:
		return []User{}, Warnings{}, fmt.Errorf("Not a valid role, it must be one of OrgManager, BillingManager and OrgAuditor")
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   Params{"organization_guid": guid},
	})
	if err != nil {
		return []User{}, nil, err
	}

	var fullUsersList []User
	warnings, err := client.paginate(request, User{}, func(item interface{}) error {
		if user, ok := item.(User); ok {
			fullUsersList = append(fullUsersList, user)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   User{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullUsersList, warnings, err
}

// GetOrganizationUsers find all users for an org .
func (client *Client) GetOrganizationUsers(guid string) ([]User, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationUsersRequest,
		URIParams:   Params{"organization_guid": guid},
	})
	if err != nil {
		return []User{}, nil, err
	}

	var fullUsersList []User
	warnings, err := client.paginate(request, User{}, func(item interface{}) error {
		if user, ok := item.(User); ok {
			fullUsersList = append(fullUsersList, user)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   User{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullUsersList, warnings, err
}

// UpdateOrganizationUserByRole makes the user or client with the given UAA ID a
// member of this role in the org . (Only available: OrgManager, BillingManager and OrgAuditor)
func (client *Client) UpdateOrganizationUserByRole(role constant.UserRole, guid string, uaaID string) (Warnings, error) {
	paramUserKey := ""
	requestName := ""
	switch role {
	case constant.OrgManager:
		paramUserKey = "manager_guid"
		requestName = internal.PutOrganizationManagerRequest
	case constant.BillingManager:
		paramUserKey = "billing_manager_guid"
		requestName = internal.PutOrganizationBillingManagerRequest
	case constant.OrgAuditor:
		paramUserKey = "auditor_guid"
		requestName = internal.PutOrganizationAuditorRequest
	case constant.OrgUser:
		paramUserKey = "user_guid"
		requestName = internal.PutOrganizationUserRequest
	default:
		return Warnings{}, fmt.Errorf("Not a valid role, it must be one of OrgManager, BillingManager, OrgAuditor and OrgUser")
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   Params{"organization_guid": guid, paramUserKey: uaaID},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// DeleteOrganizationUserByRole remove the user or client with the given UAA ID a
// member of this role in the org . (Only available: OrgManager, BillingManager and OrgAuditor)
func (client *Client) DeleteOrganizationUserByRole(role constant.UserRole, guid string, uaaID string) (Warnings, error) {
	paramUserKey := ""
	requestName := ""
	switch role {
	case constant.OrgManager:
		paramUserKey = "manager_guid"
		requestName = internal.DeleteOrganizationManagerRequest
	case constant.BillingManager:
		paramUserKey = "billing_manager_guid"
		requestName = internal.DeleteOrganizationBillingManagerRequest
	case constant.OrgAuditor:
		paramUserKey = "auditor_guid"
		requestName = internal.DeleteOrganizationAuditorRequest
	case constant.OrgUser:
		paramUserKey = "user_guid"
		requestName = internal.DeleteOrganizationUserRequest
	default:
		return Warnings{}, fmt.Errorf("Not a valid role, it must be one of OrgManager, BillingManager and OrgAuditor")
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   Params{"organization_guid": guid, paramUserKey: uaaID},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// DeleteOrganizationUserByUsername revoke the user with the given username a member of
// the org.
func (client Client) DeleteOrganizationUserByUsername(orgGUID string, username string) (Warnings, error) {
	requestBody := updateOrgUserByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationUserByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// DeleteOrganizationBillingManagerByUsername revoke the user with the given username a member of
// the org.
func (client Client) DeleteOrganizationBillingManagerByUsername(orgGUID string, username string) (Warnings, error) {
	requestBody := updateOrgUserByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationBillingManagerByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// DeleteOrganizationAuditorByUsername revoke the user with the given username a member of
// the org.
func (client Client) DeleteOrganizationAuditorByUsername(orgGUID string, username string) (Warnings, error) {
	requestBody := updateOrgUserByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationAuditorByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// DeleteOrganizationManagerByUsername revoke the user with the given username a member of
// the org.
func (client Client) DeleteOrganizationManagerByUsername(orgGUID string, username string) (Warnings, error) {
	requestBody := updateOrgUserByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationManagerByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
