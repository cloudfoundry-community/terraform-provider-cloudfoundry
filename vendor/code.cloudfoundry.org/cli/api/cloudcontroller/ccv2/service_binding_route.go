package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"encoding/json"
)

// ServiceBindingRoute represents a Cloud Controller service binding route.
type ServiceBindingRoute struct {
	// GUID is the unique service binding route identifier.
	GUID string

	// Name is the name given to the service binding route.
	Name string

	// The guid of the service instance to bind
	ServiceInstanceGuid string

	// The guid of the app to bind
	AppGuid string

	// Arbitrary parameters to pass along to the service broker.
	Parameters map[string]interface{}
}

// MarshalJSON converts an service binding route into a Cloud Controller service binding route.
func (o ServiceBindingRoute) MarshalJSON() ([]byte, error) {
	ccObj := struct {
		Name                string                 `json:"name"`
		ServiceInstanceGuid string                 `json:"service_instance_guid"`
		AppGuid             string                 `json:"app_guid"`
		Parameters          map[string]interface{} `json:"parameters,omitempty"`
	}{
		Name:                o.Name,
		ServiceInstanceGuid: o.ServiceInstanceGuid,
		AppGuid:             o.AppGuid,
		Parameters:          o.Parameters,
	}

	return json.Marshal(ccObj)
}

// UnmarshalJSON helps unmarshal a Cloud Controller service binding route response.
func (o *ServiceBindingRoute) UnmarshalJSON(data []byte) error {
	var ccObj struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name                string                 `json:"name"`
			ServiceInstanceGuid string                 `json:"service_instance_guid"`
			AppGuid             string                 `json:"app_guid"`
			Parameters          map[string]interface{} `json:"parameters"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccObj)
	if err != nil {
		return err
	}

	o.Name = ccObj.Entity.Name
	o.GUID = ccObj.Metadata.GUID
	o.ServiceInstanceGuid = ccObj.Entity.ServiceInstanceGuid
	o.AppGuid = ccObj.Entity.AppGuid
	o.Parameters = ccObj.Entity.Parameters

	return nil
}

// CreateServiceBindingRoute creates a cloud controller service binding route in with the given settings.
func (client *Client) CreateServiceBindingRoute(serviceID, routeID string, params interface{}) (Warnings, error) {
	body, err := json.Marshal(struct {
		Parameters interface{} `json:"parameters"`
	}{params})
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutServiceBindingRoutesRequest,
		Body:        bytes.NewReader(body),
		URIParams: Params{
			"service_guid": serviceID,
			"route_guid":   routeID,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{
	}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// DeleteServiceBindingRoute delete a cloud controller service binding route in with the given settings.
func (client *Client) DeleteServiceBindingRoute(serviceID, routeID string) (Warnings, error) {
	warn, err := client.deleteServiceBindingRouteRequest(internal.DeleteServiceBindingRouteRequest, serviceID, routeID)
	if err != nil {
		if _, ok := err.(ccerror.ResourceNotFoundError); ok {
			return client.deleteServiceBindingRouteRequest(internal.DeleteUserProvidedServiceInstanceRoutesRequest, serviceID, routeID)
		}
		return warn, err
	}
	return warn, err
}

func (client *Client) deleteServiceBindingRouteRequest(requestName string, serviceID, routeID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams: Params{
			"service_guid": serviceID,
			"route_guid":   routeID,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{
	}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// GetServiceBindingRoute returns back a service binding route.
func (client *Client) GetServiceBindingRoutes(serviceID string) ([]Route, Warnings, error) {
	routes, warn, err := client.getServiceBindingRoutesRequest(internal.GetServiceBindingRoutesRequest, serviceID)
	if err != nil {
		if _, ok := err.(ccerror.ResourceNotFoundError); ok {
			return client.getServiceBindingRoutesRequest(internal.GetUserProvidedServiceInstanceRoutesRequest, serviceID)
		}
		return routes, warn, err
	}
	return routes, warn, err
}

func (client *Client) getServiceBindingRoutesRequest(requestName string, serviceID string) ([]Route, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams: Params{
			"service_guid": serviceID,
		},
	})
	if err != nil {
		return []Route{}, nil, err
	}

	var fullRoutesList []Route
	warnings, err := client.paginate(request, Route{}, func(item interface{}) error {
		if i, ok := item.(Route); ok {
			fullRoutesList = append(fullRoutesList, i)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Route{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullRoutesList, warnings, err
}
