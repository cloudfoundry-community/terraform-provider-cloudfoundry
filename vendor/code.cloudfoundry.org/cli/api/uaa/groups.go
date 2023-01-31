package uaa

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/uaa/internal"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Group struct {
	ID          string `json:"id"`
	Value       string `json:"value"`
	Display     string `json:"display"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`
}

func (g Group) Name() string {
	if g.Display == "" {
		return g.DisplayName
	}
	return g.Display
}

func (g Group) Id() string {
	if g.ID == "" {
		return g.Value
	}
	return g.ID
}

// GetGroups get all UAA groups.
func (client *Client) GetGroups() ([]Group, error) {
	request, err := client.newRequest(requestOptions{
		RequestName: internal.GetGroupsRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
	})
	if err != nil {
		return []Group{}, err
	}

	var usersResources struct {
		Groups []Group `json:"resources"`
	}
	response := Response{
		Result: &usersResources,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return []Group{}, err
	}
	return usersResources.Groups, err
}

// GetGroups get all UAA groups filtered by name.
func (client *Client) GetGroupsByName(name string) ([]Group, error) {
	request, err := client.newRequest(requestOptions{
		RequestName: internal.GetGroupsRequest,
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Query: url.Values{
			"filter": []string{fmt.Sprintf(`displayName Eq "%s"`, name)},
		},
	})
	if err != nil {
		return []Group{}, err
	}

	var usersResources struct {
		Groups []Group `json:"resources"`
	}
	response := Response{
		Result: &usersResources,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return []Group{}, err
	}
	return usersResources.Groups, err
}

// AddMemberByName add member to a group by its name.
func (client *Client) AddMemberByName(userGuid, origin, groupName string) error {
	groups, err := client.GetGroupsByName(groupName)
	if err != nil {
		return err
	}

	groupAddReq := struct {
		Origin string `json:"origin"`
		Type   string `json:"type"`
		Value  string `json:"value"`
	}{
		Origin: origin,
		Value:  userGuid,
		Type:   "USER",
	}

	bodyBytes, err := json.Marshal(groupAddReq)
	if err != nil {
		return err
	}

	for _, g := range groups {
		request, err := client.newRequest(requestOptions{
			RequestName: internal.PostGroupMemberRequest,
			Header: http.Header{
				"Content-Type": {"application/json"},
			},
			URIParams: internal.Params{"group_guid": g.ID},
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
		return err
	}
	return nil
}

// DeleteMemberByName delete member to a group by its name.
func (client *Client) DeleteMemberByName(userGuid, groupName string) error {
	groups, err := client.GetGroupsByName(groupName)
	if err != nil {
		return err
	}
	for _, g := range groups {
		request, err := client.newRequest(requestOptions{
			RequestName: internal.DeleteGroupMemberRequest,
			Header: http.Header{
				"Content-Type": {"application/json"},
			},
			URIParams: internal.Params{"user_guid": userGuid, "group_guid": g.ID},
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
		return err
	}
	return nil
}
