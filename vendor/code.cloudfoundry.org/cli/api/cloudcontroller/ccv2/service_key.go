package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceKey represents a Cloud Controller Service Key.
type ServiceKey struct {
	// GUID is the unique Service Key identifier.
	GUID string
	// Name is the name of the service key.
	Name string
	// ServiceInstanceGUID is the associated service instance GUID.
	ServiceInstanceGUID string
	// Credentials are the credentials returned by the service broker for the service key.
	Credentials map[string]interface{}
}

func (serviceKey *ServiceKey) UnmarshalJSON(data []byte) error {
	var ccServiceKey struct {
		Metadata internal.Metadata
		Entity   struct {
			ServiceInstanceGUID string                 `json:"service_instance_guid"`
			Name                string                 `json:"name"`
			Credentials         map[string]interface{} `json:"credentials"`
		}
	}
	err := cloudcontroller.DecodeJSON(data, &ccServiceKey)

	if err != nil {
		return err
	}

	serviceKey.GUID = ccServiceKey.Metadata.GUID
	serviceKey.Name = ccServiceKey.Entity.Name
	serviceKey.ServiceInstanceGUID = ccServiceKey.Entity.ServiceInstanceGUID
	serviceKey.Credentials = ccServiceKey.Entity.Credentials

	return nil
}

// serviceKeyRequestBody represents the body of the service key create
// request.
type serviceKeyRequestBody struct {
	ServiceInstanceGUID string                 `json:"service_instance_guid"`
	Name                string                 `json:"name"`
	Parameters          map[string]interface{} `json:"parameters,omitempty"`
}

// CreateServiceKey creates a new service key using the provided name and
// parameters for the requested service instance.
func (client *Client) CreateServiceKey(serviceInstanceGUID string, keyName string, parameters map[string]interface{}) (ServiceKey, Warnings, error) {
	requestBody := serviceKeyRequestBody{
		ServiceInstanceGUID: serviceInstanceGUID,
		Name:                keyName,
		Parameters:          parameters,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return ServiceKey{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostServiceKeyRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return ServiceKey{}, nil, err
	}

	var serviceKey ServiceKey
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &serviceKey,
	}
	err = client.connection.Make(request, &response)

	return serviceKey, response.Warnings, err
}

// GetServiceKey returns back a service key.
func (client *Client) GetServiceKey(guid string) (ServiceKey, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceKeyRequest,
		URIParams: Params{
			"service_key_guid": guid,
		},
	})
	if err != nil {
		return ServiceKey{}, nil, err
	}

	var obj ServiceKey
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &obj,
	}

	err = client.connection.Make(request, &response)
	return obj, response.Warnings, err
}

// GetServiceKeys returns a list of Service Key based off of the provided filters.
func (client *Client) GetServiceKeys(filters ...Filter) ([]ServiceKey, Warnings, error) {
	params := ConvertFilterParameters(filters)
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceKeysRequest,
		Query:       params,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullServiceKeysList []ServiceKey
	warnings, err := client.paginate(request, ServiceKey{}, func(item interface{}) error {
		if i, ok := item.(ServiceKey); ok {
			fullServiceKeysList = append(fullServiceKeysList, i)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceKey{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullServiceKeysList, warnings, err
}

// DeleteServiceKey delete a service key.
func (client *Client) DeleteServiceKey(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteServiceKeyRequest,
		URIParams: Params{
			"service_key_guid": guid,
		},
	})
	if err != nil {
		return nil, err
	}
	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}
