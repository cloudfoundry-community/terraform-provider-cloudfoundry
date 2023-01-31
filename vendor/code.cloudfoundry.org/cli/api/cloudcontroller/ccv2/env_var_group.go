package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"encoding/json"
)

// EnvVarGroup represents a Cloud Controller env var group.
type EnvVarGroup map[string]string

// GetEnvVarGroup returns back a env var group running.
func (client *Client) GetEnvVarGroupRunning() (EnvVarGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetConfigEnvVarGroupRunningRequest,
		URIParams:   Params{},
	})
	if err != nil {
		return EnvVarGroup{}, nil, err
	}

	var obj EnvVarGroup
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &obj,
	}

	err = client.connection.Make(request, &response)
	return obj, response.Warnings, err
}

// SetEnvVarGroupRunning set a cloud controller env var group running in with the given settings.
func (client *Client) SetEnvVarGroupRunning(envVarGroup EnvVarGroup) (Warnings, error) {
	body, err := json.Marshal(envVarGroup)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutConfigEnvVarGroupRunningRequest,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// AddEnvVarRunning add env var to current env var group running.
func (client *Client) AddEnvVarRunning(envVar EnvVarGroup) (Warnings, error) {
	envVarGroup, _, err := client.GetEnvVarGroupRunning()
	if err != nil {
		return nil, err
	}

	for k, v := range envVar {
		envVarGroup[k] = v
	}

	return client.SetEnvVarGroupRunning(envVarGroup)
}

// GetEnvVarGroup returns back a env var group staging.
func (client *Client) GetEnvVarGroupStaging() (EnvVarGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetConfigEnvVarGroupStagingRequest,
		URIParams:   Params{},
	})
	if err != nil {
		return EnvVarGroup{}, nil, err
	}

	var obj EnvVarGroup
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &obj,
	}

	err = client.connection.Make(request, &response)
	return obj, response.Warnings, err
}

// SetEnvVarGroupStaging set a cloud controller env var group staging in with the given settings.
func (client *Client) SetEnvVarGroupStaging(envVarGroup EnvVarGroup) (Warnings, error) {
	body, err := json.Marshal(envVarGroup)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutConfigEnvVarGroupStagingRequest,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// AddEnvVarStaging add env var to current env var group staging.
func (client *Client) AddEnvVarStaging(envVar EnvVarGroup) (Warnings, error) {
	envVarGroup, _, err := client.GetEnvVarGroupStaging()
	if err != nil {
		return nil, err
	}

	for k, v := range envVar {
		envVarGroup[k] = v
	}

	return client.SetEnvVarGroupStaging(envVarGroup)
}
