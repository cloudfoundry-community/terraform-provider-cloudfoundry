package uaa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// User represents an UAA user account.
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"userName,omitempty"`
	Password string   `json:"password,omitempty"`
	Origin   string   `json:"origin,omitempty"`
	Name     UserName `json:"name"`
	Emails   []Email  `json:"emails"`
	Groups   []Group  `json:"groups,omitempty"`
}

type UserName struct {
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
}

type Email struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}

// CreateUser creates a new UAA user account with the provided password.
func (client *Client) CreateUser(user string, password string, origin string) (User, error) {
	userRequest := User{
		Username: user,
		Password: password,
		Origin:   origin,
		Name: UserName{
			FamilyName: user,
			GivenName:  user,
		},
		Emails: []Email{
			{
				Value:   user,
				Primary: true,
			},
		},
	}

	bodyBytes, err := json.Marshal(userRequest)
	if err != nil {
		return User{}, err
	}

	request, err := client.newRequest(requestOptions{
		RequestName: internal.PostUserRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Body: bytes.NewBuffer(bodyBytes),
	})
	if err != nil {
		return User{}, err
	}

	var userResponse User
	response := Response{
		Result: &userResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return User{}, err
	}

	return User(userResponse), nil
}

// CreateUser creates a new UAA user account with the provided object.
func (client *Client) CreateUserFromObject(user User) (User, error) {
	bodyBytes, err := json.Marshal(user)
	if err != nil {
		return User{}, err
	}
	request, err := client.newRequest(requestOptions{
		RequestName: internal.PostUserRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Body: bytes.NewBuffer(bodyBytes),
	})
	if err != nil {
		return User{}, err
	}

	var userResponse User
	response := Response{
		Result: &userResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return User{}, err
	}

	return User(userResponse), nil
}

// DeleteUser delete an UAA user account.
func (client *Client) DeleteUser(guid string) error {
	request, err := client.newRequest(requestOptions{
		RequestName: internal.DeleteUserRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		URIParams: internal.Params{"user_guid": guid},
	})
	if err != nil {
		return err
	}

	var userResponse User
	response := Response{
		Result: &userResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return err
	}

	return nil
}

// GetUser get an UAA user account by its id.
func (client *Client) GetUser(guid string) (User, error) {
	request, err := client.newRequest(requestOptions{
		RequestName: internal.GetUserRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		URIParams: internal.Params{"user_guid": guid},
	})
	if err != nil {
		return User{}, err
	}

	var userResponse User
	response := Response{
		Result: &userResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return User{}, err
	}

	return User(userResponse), nil
}

// UpdateUser update a UAA user account.
func (client *Client) UpdateUser(user User) (User, error) {
	bodyBytes, err := json.Marshal(user)
	if err != nil {
		return User{}, err
	}
	fmt.Println(string(bodyBytes))
	request, err := client.newRequest(requestOptions{
		RequestName: internal.PutUserRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
			"If-Match":     {"*"},
		},
		Body:      bytes.NewBuffer(bodyBytes),
		URIParams: internal.Params{"user_guid": user.ID},
	})
	if err != nil {
		return User{}, err
	}

	var userResponse User
	response := Response{
		Result: &userResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return User{}, err
	}

	return User(userResponse), nil
}

// GetUsers get all UAA user account by its username.
func (client *Client) GetUsersByUsername(username string) ([]User, error) {
	request, err := client.newRequest(requestOptions{
		RequestName: internal.GetUsersRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Query: url.Values{
			"attributes": []string{"id,userName"},
			"filter":     []string{fmt.Sprintf(`userName Eq "%s"`, username)},
		},
	})
	if err != nil {
		return []User{}, err
	}

	var usersResources struct {
		Users []User `json:"resources"`
	}
	response := Response{
		Result: &usersResources,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return []User{}, err
	}
	return usersResources.Users, err
}

// ChangeUserPassword change an user password by its id.
func (client *Client) ChangeUserPassword(guid, oldPass, newPass string) error {
	changePassRequest := struct {
		OldPassword string `json:"oldPassword"`
		Password    string `json:"password"`
	}{
		OldPassword: oldPass,
		Password:    newPass,
	}
	bodyBytes, err := json.Marshal(changePassRequest)
	if err != nil {
		return err
	}
	request, err := client.newRequest(requestOptions{
		RequestName: internal.PutUserPasswordRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		URIParams: internal.Params{"user_guid": guid},
		Body:      bytes.NewBuffer(bodyBytes),
	})
	if err != nil {
		return err
	}

	response := Response{
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return err
	}

	return nil
}
