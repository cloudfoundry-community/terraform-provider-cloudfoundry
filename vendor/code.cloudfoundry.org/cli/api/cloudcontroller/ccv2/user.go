package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// User represents a Cloud Controller User.
type User struct {
	// GUID is the unique user identifier.
	GUID string

	// Username of the user
	Username string
}

// UnmarshalJSON helps unmarshal a Cloud Controller User response.
func (user *User) UnmarshalJSON(data []byte) error {
	var ccUser struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Username string `json:"username"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccUser)
	if err != nil {
		return err
	}

	user.GUID = ccUser.Metadata.GUID
	user.Username = ccUser.Entity.Username
	return nil
}

// CreateUser creates a new Cloud Controller User from the provided UAA user
// ID.
func (client *Client) CreateUser(uaaUserID string) (User, Warnings, error) {
	type userRequestBody struct {
		GUID string `json:"guid"`
	}

	bodyBytes, err := json.Marshal(userRequestBody{
		GUID: uaaUserID,
	})
	if err != nil {
		return User{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostUserRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return User{}, nil, err
	}

	var user User
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &user,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return User{}, response.Warnings, err
	}

	return user, response.Warnings, nil
}

// GetUserOrganizations get all organizations available to user
func (client *Client) GetUserOrganizations(uaaUserID string, filters ...Filter) ([]Organization, Warnings, error) {
	allQueries := ConvertFilterParameters(filters)
	allQueries.Add("order-by", "name")
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUserOrganizationsRequest,
		URIParams:   Params{"user_guid": uaaUserID},
		Query:       allQueries,
	})
	if err != nil {
		return []Organization{}, nil, err
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

// GetUserSpaces get all spaces available to user
func (client *Client) GetUserSpaces(uaaUserID string, filters ...Filter) ([]Space, Warnings, error) {
	allQueries := ConvertFilterParameters(filters)
	allQueries.Add("order-by", "name")
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUserSpacesRequest,
		URIParams:   Params{"user_guid": uaaUserID},
		Query:       allQueries,
	})
	if err != nil {
		return []Space{}, nil, err
	}

	var fullSpacesList []Space
	warnings, err := client.paginate(request, Space{}, func(item interface{}) error {
		if org, ok := item.(Space); ok {
			fullSpacesList = append(fullSpacesList, org)
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

// GetUsers returns back a list of User based off of the
// provided filters.
func (client *Client) GetUsers(filters ...Filter) ([]User, Warnings, error) {
	allQueries := ConvertFilterParameters(filters)
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUsersRequest,
		Query:       allQueries,
	})

	if err != nil {
		return nil, nil, err
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
