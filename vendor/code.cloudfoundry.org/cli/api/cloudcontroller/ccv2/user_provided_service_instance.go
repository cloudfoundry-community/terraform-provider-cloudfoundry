package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"encoding/json"
)

// UserProvidedServiceInstance represents a Cloud Controller user provided service instance.
type UserProvidedServiceInstance struct {
	// GUID is the unique user provided service instance identifier.
	GUID string

	// Name is the name given to the user provided service instance.
	Name string

	// The guid of the space in which the instance will be created
	SpaceGuid string

	// URL to which logs will be streamed for bound applications.
	SyslogDrainUrl string

	// A hash exposed in the VCAP_SERVICES environment variable for bound applications.
	Credentials map[string]interface{}

	// URL to which requests for bound routes will be forwarded.
	RouteServiceUrl string

	// A list of tags for the service instance. Max characters: 2048
	Tags []string
}

// MarshalJSON converts an user provided service instance into a Cloud Controller user provided service instance.
func (o UserProvidedServiceInstance) MarshalJSON() ([]byte, error) {
	ccObj := struct {
		Name            string                 `json:"name,omitempty"`
		SpaceGuid       string                 `json:"space_guid,omitempty"`
		SyslogDrainUrl  string                 `json:"syslog_drain_url,omitempty"`
		Credentials     map[string]interface{} `json:"credentials,omitempty"`
		RouteServiceUrl string                 `json:"route_service_url,omitempty"`
		Tags            []string               `json:"tags,omitempty"`
	}{
		Name:            o.Name,
		SpaceGuid:       o.SpaceGuid,
		SyslogDrainUrl:  o.SyslogDrainUrl,
		Credentials:     o.Credentials,
		RouteServiceUrl: o.RouteServiceUrl,
		Tags:            o.Tags,
	}

	return json.Marshal(ccObj)
}

// UnmarshalJSON helps unmarshal a Cloud Controller user provided service instance response.
func (o *UserProvidedServiceInstance) UnmarshalJSON(data []byte) error {
	var ccObj struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name            string                 `json:"name"`
			SpaceGuid       string                 `json:"space_guid"`
			SyslogDrainUrl  string                 `json:"syslog_drain_url"`
			Credentials     map[string]interface{} `json:"credentials"`
			RouteServiceUrl string                 `json:"route_service_url"`
			Tags            []string               `json:"tags"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccObj)
	if err != nil {
		return err
	}

	o.Name = ccObj.Entity.Name
	o.GUID = ccObj.Metadata.GUID
	o.SpaceGuid = ccObj.Entity.SpaceGuid
	o.SyslogDrainUrl = ccObj.Entity.SyslogDrainUrl
	o.Credentials = ccObj.Entity.Credentials
	o.RouteServiceUrl = ccObj.Entity.RouteServiceUrl
	o.Tags = ccObj.Entity.Tags

	return nil
}

// GetUserProvidedServiceInstances returns back a list of user provided service instances based off of the
// provided filters.
func (client *Client) GetUserProvServiceInstances(filters ...Filter) ([]UserProvidedServiceInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUserProvidedServiceInstancesRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullObjList []UserProvidedServiceInstance
	warnings, err := client.paginate(request, UserProvidedServiceInstance{}, func(item interface{}) error {
		if app, ok := item.(UserProvidedServiceInstance); ok {
			fullObjList = append(fullObjList, app)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   UserProvidedServiceInstance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullObjList, warnings, err
}

// GetUserProvidedServiceInstances returns back a list of *user provided* Service Instances based
// off the provided queries.
func (client *Client) GetUserProvidedServiceInstances(filters ...Filter) ([]ServiceInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUserProvidedServiceInstancesRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullInstancesList []ServiceInstance
	warnings, err := client.paginate(request, ServiceInstance{}, func(item interface{}) error {
		if instance, ok := item.(ServiceInstance); ok {
			fullInstancesList = append(fullInstancesList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceInstance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullInstancesList, warnings, err
}

// CreateUserProvidedServiceInstance creates a cloud controller user provided service instance in with the given settings.
func (client *Client) CreateUserProvidedServiceInstance(userProvidedServiceInstance UserProvidedServiceInstance) (UserProvidedServiceInstance, Warnings, error) {
	body, err := json.Marshal(userProvidedServiceInstance)
	if err != nil {
		return UserProvidedServiceInstance{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostUserProvidedServiceInstancesRequest,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return UserProvidedServiceInstance{}, nil, err
	}

	var updatedObj UserProvidedServiceInstance
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}

// UpdateUserProvidedServiceInstance updates the user provided service instance with the given GUID.
func (client *Client) UpdateUserProvidedServiceInstance(userProvidedServiceInstance UserProvidedServiceInstance) (UserProvidedServiceInstance, Warnings, error) {
	body, err := json.Marshal(userProvidedServiceInstance)
	if err != nil {
		return UserProvidedServiceInstance{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutUserProvidedServiceInstanceRequest,
		URIParams:   Params{"user_provided_service_instance_guid": userProvidedServiceInstance.GUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return UserProvidedServiceInstance{}, nil, err
	}

	var updatedObj UserProvidedServiceInstance
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}

// GetUserProvidedServiceInstance returns back a user provided service instance.
func (client *Client) GetUserProvidedServiceInstance(guid string) (UserProvidedServiceInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUserProvidedServiceInstanceRequest,
		URIParams: Params{
			"user_provided_service_instance_guid": guid,
		},
	})
	if err != nil {
		return UserProvidedServiceInstance{}, nil, err
	}

	var obj UserProvidedServiceInstance
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &obj,
	}

	err = client.connection.Make(request, &response)
	return obj, response.Warnings, err
}

// DeleteUserProvidedServiceInstance delete a user provided service instance
func (client *Client) DeleteUserProvidedServiceInstance(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteUserProvidedServiceInstanceRequest,
		URIParams: Params{
			"user_provided_service_instance_guid": guid,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}
