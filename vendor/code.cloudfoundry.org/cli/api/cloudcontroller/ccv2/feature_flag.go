package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"encoding/json"
)

// FeatureFlag represents a Cloud Controller feature flag.
type FeatureFlag struct {
	// Name is a string representation of the Cloud Controller
	// feature flag's name.
	Name string `json:"name"`

	// Enabled is the status of the Cloud Controller feature
	// flag.
	Enabled bool `json:"enabled"`
}

// GetConfigFeatureFlags retrieves a list of FeatureFlag from the Cloud
// Controller.
func (client Client) GetConfigFeatureFlags() ([]FeatureFlag, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetConfigFeatureFlagsRequest,
	})
	if err != nil {
		return nil, nil, err
	}

	var featureFlags []FeatureFlag
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &featureFlags,
	}

	err = client.connection.Make(request, &response)
	return featureFlags, response.Warnings, err
}

// GetConfigFeatureFlags retrieves a list of FeatureFlag from the Cloud
// Controller.
func (client Client) SetConfigFeatureFlags(ff FeatureFlag) (Warnings, error) {
	body, err := json.Marshal(struct {
		Enabled bool `json:"enabled"`
	}{
		Enabled: ff.Enabled,
	})
	if err != nil {
		return nil, err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutConfigFeatureFlagsRequest,
		URIParams:   Params{"name": ff.Name},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{
	}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}
