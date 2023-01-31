package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"encoding/json"
)

// SecurityGroup represents a Cloud Controller Security Group.
type SecurityGroup struct {
	// GUID is the unique Security Group identifier.
	GUID string
	// Name is the Security Group's name.
	Name string
	// Rules are the Security Group Rules associated with this Security Group.
	Rules []SecurityGroupRule
	// RunningDefault is true when this Security Group is applied to all running
	// apps in the CF instance.
	RunningDefault bool
	// StagingDefault is true when this Security Group is applied to all staging
	// apps in the CF instance.
	StagingDefault bool
}

// UnmarshalJSON helps unmarshal a Cloud Controller Security Group response
func (securityGroup *SecurityGroup) UnmarshalJSON(data []byte) error {
	var ccSecurityGroup struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			GUID  string `json:"guid"`
			Name  string `json:"name"`
			Rules []struct {
				Description string `json:"description"`
				Destination string `json:"destination"`
				Ports       string `json:"ports"`
				Protocol    string `json:"protocol"`
				Log         *bool  `json:"log"`
				Code        *int   `json:"code"`
				Type        *int   `json:"type"`
			} `json:"rules"`
			RunningDefault bool `json:"running_default"`
			StagingDefault bool `json:"staging_default"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccSecurityGroup)
	if err != nil {
		return err
	}

	securityGroup.GUID = ccSecurityGroup.Metadata.GUID
	securityGroup.Name = ccSecurityGroup.Entity.Name
	securityGroup.Rules = make([]SecurityGroupRule, len(ccSecurityGroup.Entity.Rules))
	for i, ccRule := range ccSecurityGroup.Entity.Rules {
		securityGroup.Rules[i].Description = ccRule.Description
		securityGroup.Rules[i].Destination = ccRule.Destination
		securityGroup.Rules[i].Ports = ccRule.Ports
		securityGroup.Rules[i].Protocol = ccRule.Protocol
		securityGroup.Rules[i].Log.ParseBoolValue(ccRule.Log)
		securityGroup.Rules[i].Type.ParseIntValue(ccRule.Type)
		securityGroup.Rules[i].Code.ParseIntValue(ccRule.Code)
	}
	securityGroup.RunningDefault = ccSecurityGroup.Entity.RunningDefault
	securityGroup.StagingDefault = ccSecurityGroup.Entity.StagingDefault
	return nil
}

// MarshalJSON helps marshal a Cloud Controller Security Group request
func (securityGroup SecurityGroup) MarshalJSON() ([]byte, error) {
	type rule struct {
		Description string `json:"description,omitempty"`
		Destination string `json:"destination"`
		Ports       string `json:"ports,omitempty"`
		Protocol    string `json:"protocol"`
		Log         bool   `json:"log"`
		Code        *int   `json:"code,omitempty"`
		Type        *int   `json:"type,omitempty"`
	}
	ccObj := struct {
		Name  string `json:"name,omitempty"`
		Rules []rule `json:"rules,omitempty"`
	}{
		Name:  securityGroup.Name,
		Rules: make([]rule, 0),
	}
	for _, ccRule := range securityGroup.Rules {
		r := rule{
			Protocol:    ccRule.Protocol,
			Description: ccRule.Description,
			Destination: ccRule.Destination,
			Ports:       ccRule.Ports,
		}
		if ccRule.Log.IsSet {
			r.Log = ccRule.Log.Value
		}
		if ccRule.Code.IsSet {
			r.Code = &ccRule.Code.Value
		}
		if ccRule.Type.IsSet {
			r.Type = &ccRule.Type.Value
		}
		ccObj.Rules = append(ccObj.Rules, r)
	}

	return json.Marshal(ccObj)
}

// DeleteSecurityGroupSpace disassociates a security group in the running phase
// for the lifecycle, specified by its GUID, from a space, which is also
// specified by its GUID.
func (client *Client) DeleteSecurityGroupSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSecurityGroupSpaceRequest,
		URIParams: Params{
			"security_group_guid": securityGroupGUID,
			"space_guid":          spaceGUID,
		},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// DeleteSecurityGroupStagingSpace disassociates a security group in the
// staging phase fo the lifecycle, specified by its GUID, from a space, which
// is also specified by its GUID.
func (client *Client) DeleteSecurityGroupStagingSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSecurityGroupStagingSpaceRequest,
		URIParams: Params{
			"security_group_guid": securityGroupGUID,
			"space_guid":          spaceGUID,
		},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// GetSecurityGroups returns a list of Security Groups based off the provided
// filters.
func (client *Client) GetSecurityGroups(filters ...Filter) ([]SecurityGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSecurityGroupsRequest,
		Query:       ConvertFilterParameters(filters),
	})

	if err != nil {
		return nil, nil, err
	}

	var securityGroupsList []SecurityGroup
	warnings, err := client.paginate(request, SecurityGroup{}, func(item interface{}) error {
		if securityGroup, ok := item.(SecurityGroup); ok {
			securityGroupsList = append(securityGroupsList, securityGroup)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   SecurityGroup{},
				Unexpected: item,
			}
		}
		return nil
	})

	return securityGroupsList, warnings, err
}

// GetSpaceSecurityGroups returns the running Security Groups associated with
// the provided Space GUID.
func (client *Client) GetSpaceSecurityGroups(spaceGUID string, filters ...Filter) ([]SecurityGroup, Warnings, error) {
	return client.getSpaceSecurityGroupsBySpaceAndLifecycle(spaceGUID, internal.GetSpaceSecurityGroupsRequest, filters)
}

// GetSpaceStagingSecurityGroups returns the staging Security Groups
// associated with the provided Space GUID.
func (client *Client) GetSpaceStagingSecurityGroups(spaceGUID string, filters ...Filter) ([]SecurityGroup, Warnings, error) {
	return client.getSpaceSecurityGroupsBySpaceAndLifecycle(spaceGUID, internal.GetSpaceStagingSecurityGroupsRequest, filters)
}

// UpdateSecurityGroupSpace associates a security group in the running phase
// for the lifecycle, specified by its GUID, from a space, which is also
// specified by its GUID.
func (client *Client) UpdateSecurityGroupSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSecurityGroupSpaceRequest,
		URIParams: Params{
			"security_group_guid": securityGroupGUID,
			"space_guid":          spaceGUID,
		},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// UpdateSecurityGroupStagingSpace associates a security group in the staging
// phase for the lifecycle, specified by its GUID, from a space, which is also
// specified by its GUID.
func (client *Client) UpdateSecurityGroupStagingSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSecurityGroupStagingSpaceRequest,
		URIParams: Params{
			"security_group_guid": securityGroupGUID,
			"space_guid":          spaceGUID,
		},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

func (client *Client) getSpaceSecurityGroupsBySpaceAndLifecycle(spaceGUID string, lifecycle string, filters []Filter) ([]SecurityGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: lifecycle,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var securityGroupsList []SecurityGroup
	warnings, err := client.paginate(request, SecurityGroup{}, func(item interface{}) error {
		if securityGroup, ok := item.(SecurityGroup); ok {
			securityGroupsList = append(securityGroupsList, securityGroup)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   SecurityGroup{},
				Unexpected: item,
			}
		}
		return err
	})

	return securityGroupsList, warnings, err
}

// GetRunningSecurityGroups returns back a list of running security groups based off of the
// provided filters.
func (client *Client) GetRunningSecurityGroups(filters ...Filter) ([]SecurityGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetConfigRunningSecurityGroupsRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullObjList []SecurityGroup
	warnings, err := client.paginate(request, SecurityGroup{}, func(item interface{}) error {
		if app, ok := item.(SecurityGroup); ok {
			fullObjList = append(fullObjList, app)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   SecurityGroup{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullObjList, warnings, err
}

// BindSecurityGroup bind the security group with the given GUID.
func (client *Client) BindRunningSecurityGroup(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutConfigRunningSecurityGroupRequest,
		URIParams:   Params{"security_group_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{
	}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// UnbindSecurityGroup delete a security group
func (client *Client) UnbindRunningSecurityGroup(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteConfigRunningSecurityGroupRequest,
		URIParams: Params{
			"security_group_guid": guid,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// GetSecurityGroups returns back a list of staging security groups based off of the
// provided filters.
func (client *Client) GetStagingSecurityGroups(filters ...Filter) ([]SecurityGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetConfigStagingSecurityGroupsRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullObjList []SecurityGroup
	warnings, err := client.paginate(request, SecurityGroup{}, func(item interface{}) error {
		if app, ok := item.(SecurityGroup); ok {
			fullObjList = append(fullObjList, app)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   SecurityGroup{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullObjList, warnings, err
}

// BindSecurityGroup bind the security group with the given GUID.
func (client *Client) BindStagingSecurityGroup(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutConfigStagingSecurityGroupRequest,
		URIParams:   Params{"security_group_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{
	}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// UnbindSecurityGroup delete a security group
func (client *Client) UnbindStagingSecurityGroup(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteConfigStagingSecurityGroupRequest,
		URIParams: Params{
			"security_group_guid": guid,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// CreateSecurityGroup creates a cloud controller security group in with the given settings.
func (client *Client) CreateSecurityGroup(securityGroup SecurityGroup) (SecurityGroup, Warnings, error) {
	body, err := json.Marshal(securityGroup)
	if err != nil {
		return SecurityGroup{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSecurityGroupsRequest,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return SecurityGroup{}, nil, err
	}

	var updatedObj SecurityGroup
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}

// UpdateSecurityGroup updates the security group with the given GUID.
func (client *Client) UpdateSecurityGroup(securityGroup SecurityGroup) (SecurityGroup, Warnings, error) {
	body, err := json.Marshal(securityGroup)
	if err != nil {
		return SecurityGroup{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSecurityGroupRequest,
		URIParams:   Params{"security_group_guid": securityGroup.GUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return SecurityGroup{}, nil, err
	}

	var updatedObj SecurityGroup
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}

// GetSecurityGroup returns back a security group.
func (client *Client) GetSecurityGroup(guid string) (SecurityGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSecurityGroupRequest,
		URIParams: Params{
			"security_group_guid": guid,
		},
	})
	if err != nil {
		return SecurityGroup{}, nil, err
	}

	var obj SecurityGroup
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &obj,
	}

	err = client.connection.Make(request, &response)
	return obj, response.Warnings, err
}

// DeleteSecurityGroup delete a security group
func (client *Client) DeleteSecurityGroup(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSecurityGroupRequest,
		URIParams: Params{
			"security_group_guid": guid,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}
