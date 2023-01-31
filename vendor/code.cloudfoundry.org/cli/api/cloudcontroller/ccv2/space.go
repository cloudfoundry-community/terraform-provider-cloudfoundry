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

// Space represents a Cloud Controller Space.
type Space struct {
	// AllowSSH specifies whether SSH is enabled for this space.
	AllowSSH bool

	// GUID is the unique space identifier.
	GUID string

	// Name is the name given to the space.
	Name string

	// OrganizationGUID is the unique identifier of the organization this space
	// belongs to.
	OrganizationGUID string

	// SpaceQuotaDefinitionGUID is the unique identifier of the space quota
	// defined for this space.
	SpaceQuotaDefinitionGUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Space response.
func (space *Space) UnmarshalJSON(data []byte) error {
	var ccSpace struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name                     string `json:"name"`
			AllowSSH                 bool   `json:"allow_ssh"`
			SpaceQuotaDefinitionGUID string `json:"space_quota_definition_guid"`
			OrganizationGUID         string `json:"organization_guid"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccSpace)
	if err != nil {
		return err
	}

	space.GUID = ccSpace.Metadata.GUID
	space.Name = ccSpace.Entity.Name
	space.AllowSSH = ccSpace.Entity.AllowSSH
	space.SpaceQuotaDefinitionGUID = ccSpace.Entity.SpaceQuotaDefinitionGUID
	space.OrganizationGUID = ccSpace.Entity.OrganizationGUID
	return nil
}

// UnmarshalJSON helps unmarshal a Cloud Controller Space response.
func (space Space) MarshalJSON() ([]byte, error) {
	ccObj := struct {
		Name                     string `json:"name,omitempty"`
		OrganizationGUID         string `json:"organization_guid,omitempty"`
		AllowSSH                 bool   `json:"allow_ssh"`
		SpaceQuotaDefinitionGUID string `json:"space_quota_definition_guid,omitempty"`
	}{
		Name:                     space.Name,
		OrganizationGUID:         space.OrganizationGUID,
		AllowSSH:                 space.AllowSSH,
		SpaceQuotaDefinitionGUID: space.SpaceQuotaDefinitionGUID,
	}

	return json.Marshal(ccObj)
}

type createSpaceRequestBody struct {
	Name             string `json:"name"`
	OrganizationGUID string `json:"organization_guid"`
}

// CreateSpace creates a new space with the provided spaceName in the org with
// the provided orgGUID.
func (client *Client) CreateSpace(spaceName string, orgGUID string) (Space, Warnings, error) {
	requestBody := createSpaceRequestBody{
		Name:             spaceName,
		OrganizationGUID: orgGUID,
	}

	bodyBytes, _ := json.Marshal(requestBody)

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSpaceRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Space{}, nil, err
	}

	var space Space
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &space,
	}

	err = client.connection.Make(request, &response)

	return space, response.Warnings, err
}

// CreateSpace creates a new space with the provided spaceName in the org with
// the provided orgGUID.
func (client *Client) CreateSpaceFromObject(space Space) (Space, Warnings, error) {
	bodyBytes, _ := json.Marshal(space)

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSpaceRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Space{}, nil, err
	}

	var updateSpace Space
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updateSpace,
	}

	err = client.connection.Make(request, &response)

	return updateSpace, response.Warnings, err
}

// DeleteSpace deletes the Space associated with the provided
// GUID. It will return the Cloud Controller job that is assigned to the
// Space deletion.
func (client *Client) DeleteSpace(guid string) (Job, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSpaceRequest,
		URIParams:   Params{"space_guid": guid},
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

// GetSecurityGroupSpaces returns a list of Spaces based on the provided
// SecurityGroup GUID.
func (client *Client) GetSecurityGroupSpaces(securityGroupGUID string) ([]Space, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSecurityGroupSpacesRequest,
		URIParams:   map[string]string{"security_group_guid": securityGroupGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullSpacesList []Space
	warnings, err := client.paginate(request, Space{}, func(item interface{}) error {
		if space, ok := item.(Space); ok {
			fullSpacesList = append(fullSpacesList, space)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Space{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSpacesList, warnings, err
}

// GetSecurityGroupStagingSpaces returns a list of Spaces based on the provided
// SecurityGroup GUID.
func (client *Client) GetSecurityGroupStagingSpaces(securityGroupGUID string) ([]Space, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSecurityGroupStagingSpacesRequest,
		URIParams:   map[string]string{"security_group_guid": securityGroupGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullSpacesList []Space
	warnings, err := client.paginate(request, Space{}, func(item interface{}) error {
		if space, ok := item.(Space); ok {
			fullSpacesList = append(fullSpacesList, space)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Space{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSpacesList, warnings, err
}

// GetSpaces returns a list of Spaces based off of the provided filters.
func (client *Client) GetSpaces(filters ...Filter) ([]Space, Warnings, error) {
	params := ConvertFilterParameters(filters)
	params.Add("order-by", "name")
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpacesRequest,
		Query:       params,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullSpacesList []Space
	warnings, err := client.paginate(request, Space{}, func(item interface{}) error {
		if space, ok := item.(Space); ok {
			fullSpacesList = append(fullSpacesList, space)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Space{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSpacesList, warnings, err
}

// UpdateSpaceDeveloper grants the space developer role to the user or client
// associated with the given UAA ID.
func (client *Client) UpdateSpaceDeveloper(spaceGUID string, uaaID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceDeveloperRequest,
		URIParams: map[string]string{
			"space_guid":     spaceGUID,
			"developer_guid": uaaID,
		},
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

type updateRoleRequestBody struct {
	Username string `json:"username"`
}

// UpdateSpaceDeveloperByUsername grants the given username the space developer role.
func (client *Client) UpdateSpaceDeveloperByUsername(spaceGUID string, username string) (Warnings, error) {
	requestBody := updateRoleRequestBody{
		Username: username,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceDeveloperByUsernameRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return Warnings(response.Warnings), err
}

// UpdateSpaceAuditorByUsername grants the given username the space developer role.
func (client *Client) UpdateSpaceAuditorByUsername(spaceGUID string, username string) (Warnings, error) {
	requestBody := updateRoleRequestBody{
		Username: username,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceAuditorByUsernameRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return Warnings(response.Warnings), err
}

// UpdateSpaceManager grants the space manager role to the user or client
// associated with the given UAA ID.
func (client *Client) UpdateSpaceManager(spaceGUID string, uaaID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceManagerRequest,
		URIParams: map[string]string{
			"space_guid":   spaceGUID,
			"manager_guid": uaaID,
		},
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// UpdateSpaceManagerByUsername grants the given username the space manager role.
func (client *Client) UpdateSpaceManagerByUsername(spaceGUID string, username string) (Warnings, error) {
	requestBody := updateRoleRequestBody{
		Username: username,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceManagerByUsernameRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// GetSpace returns back a space.
func (client *Client) GetSpace(guid string) (Space, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceRequest,
		URIParams: Params{
			"space_guid": guid,
		},
	})
	if err != nil {
		return Space{}, nil, err
	}

	var obj Space
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &obj,
	}

	err = client.connection.Make(request, &response)
	return obj, response.Warnings, err
}

// UpdateSpace updates the space with the given GUID.
func (client *Client) UpdateSpace(space Space) (Space, Warnings, error) {
	body, err := json.Marshal(space)
	if err != nil {
		return Space{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceRequest,
		URIParams:   Params{"space_guid": space.GUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Space{}, nil, err
	}

	var updatedObj Space
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}

// UpdateSpaceUserByRole makes the user or client with the given UAA ID a
// member of this role in the space . (Only available: SpaceManager, SpaceDeveloper and SpaceAuditor)
func (client *Client) UpdateSpaceUserByRole(role constant.UserRole, guid string, uaaID string) (Warnings, error) {
	paramUserKey := ""
	requestName := ""
	switch role {
	case constant.SpaceManager:
		paramUserKey = "manager_guid"
		requestName = internal.PutSpaceManagerRequest
	case constant.SpaceDeveloper:
		paramUserKey = "developer_guid"
		requestName = internal.PutSpaceDeveloperRequest
	case constant.SpaceAuditor:
		paramUserKey = "auditor_guid"
		requestName = internal.PutSpaceAuditorRequest
	default:
		return Warnings{}, fmt.Errorf("Not a valid role, it must be one of SpaceManager, SpaceDeveloper and SpaceAuditor")
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   Params{"space_guid": guid, paramUserKey: uaaID},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// UpdateSpaceUserByRole makes the user or client with the given UAA ID a
// member of this role in the space . (Only available: SpaceManager, SpaceDeveloper and SpaceAuditor)
func (client *Client) DeleteSpaceUserByRole(role constant.UserRole, guid string, uaaID string) (Warnings, error) {
	paramUserKey := ""
	requestName := ""
	switch role {
	case constant.SpaceManager:
		paramUserKey = "manager_guid"
		requestName = internal.DeleteSpaceManagerRequest
	case constant.SpaceDeveloper:
		paramUserKey = "developer_guid"
		requestName = internal.DeleteSpaceDeveloperRequest
	case constant.SpaceAuditor:
		paramUserKey = "auditor_guid"
		requestName = internal.DeleteSpaceAuditorRequest
	default:
		return Warnings{}, fmt.Errorf("Not a valid role, it must be one of SpaceManager, SpaceDeveloper and SpaceAuditor")
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   Params{"space_guid": guid, paramUserKey: uaaID},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// GetSpaceUsersByRole find all users for a space by role .
// (Only available: SpaceManager, SpaceDeveloper and SpaceAuditor)
func (client *Client) GetSpaceUsersByRole(role constant.UserRole, guid string) ([]User, Warnings, error) {
	requestName := ""
	switch role {
	case constant.SpaceManager:
		requestName = internal.GetSpaceManagersRequest
	case constant.SpaceDeveloper:
		requestName = internal.GetSpaceDevelopersRequest
	case constant.SpaceAuditor:
		requestName = internal.GetSpaceAuditorsRequest
	default:
		return []User{}, Warnings{}, fmt.Errorf("Not a valid role, it must be one of SpaceManager, SpaceDeveloper and SpaceAuditor")
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   Params{"space_guid": guid},
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

// DeleteSpaceManagerByUsername revoke the given username the space manager role.
func (client *Client) DeleteSpaceManagerByUsername(spaceGUID string, username string) (Warnings, error) {
	requestBody := updateRoleRequestBody{
		Username: username,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSpaceManagerByUsernameRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// DeleteSpaceAuditorByUsername revoke the given username the space manager role.
func (client *Client) DeleteSpaceAuditorByUsername(spaceGUID string, username string) (Warnings, error) {
	requestBody := updateRoleRequestBody{
		Username: username,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSpaceAuditorByUsernameRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// DeleteSpaceDeveloperByUsername revoke the given username the space manager role.
func (client *Client) DeleteSpaceDeveloperByUsername(spaceGUID string, username string) (Warnings, error) {
	requestBody := updateRoleRequestBody{
		Username: username,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSpaceDeveloperByUsernameRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
