package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"encoding/json"
)

// GetSpaceQuotaDefinition returns a Space Quota.
func (client *Client) GetSpaceQuotaDefinition(guid string) (Quota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceQuotaDefinitionRequest,
		URIParams:   Params{"space_quota_guid": guid},
	})
	if err != nil {
		return Quota{}, nil, err
	}

	var spaceQuota Quota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &spaceQuota,
	}

	err = client.connection.Make(request, &response)
	return spaceQuota, response.Warnings, err
}

// GetSpaceQuotas returns all the space quotas for the org
func (client *Client) GetSpaceQuotas(spaceGUID string) ([]Quota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationSpaceQuotasRequest,
		URIParams:   Params{"space_guid": spaceGUID},
	})

	if err != nil {
		return nil, nil, err
	}

	var spaceQuotas []Quota
	warnings, err := client.paginate(request, Quota{}, func(item interface{}) error {
		if spaceQuota, ok := item.(Quota); ok {
			spaceQuotas = append(spaceQuotas, spaceQuota)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Quota{},
				Unexpected: item,
			}
		}
		return nil
	})

	return spaceQuotas, warnings, err
}

// SetSpaceQuota should set the quota for the space and returns the warnings
func (client *Client) SetSpaceQuota(spaceGUID string, quotaGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceQuotaRequest,
		URIParams:   Params{"space_quota_guid": quotaGUID, "space_guid": spaceGUID},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// DeleteSpaceQuota delete a space quota
func (client *Client) DeleteSpaceQuota(quotaGuid, spaceGuid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSpaceQuotaRequest,
		URIParams: Params{
			"space_quota_guid": quotaGuid,
			"space_guid":       spaceGuid,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// GetQuotas returns back a list of quotas based off of the
// provided filters.
func (client *Client) GetSpaceQuotaDefinitions(filters ...Filter) ([]Quota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceQuotaDefinitionsRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullObjList []Quota
	warnings, err := client.paginate(request, Quota{}, func(item interface{}) error {
		if app, ok := item.(Quota); ok {
			fullObjList = append(fullObjList, app)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Quota{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullObjList, warnings, err
}

// CreateQuota creates a cloud controller quota in with the given settings.
func (client *Client) CreateSpaceQuotaDefinition(quota Quota) (Quota, Warnings, error) {
	body, err := json.Marshal(quota)
	if err != nil {
		return Quota{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSpaceQuotaDefinitionsRequest,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Quota{}, nil, err
	}

	var updatedObj Quota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}

// DeleteQuota delete a quota
func (client *Client) DeleteSpaceQuotaDefinition(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSpaceQuotaDefinitionRequest,
		URIParams: Params{
			"space_quota_guid": guid,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// UpdateQuota updates the quota with the given GUID.
func (client *Client) UpdateSpaceQuotaDefinition(quota Quota) (Quota, Warnings, error) {
	body, err := json.Marshal(quota)
	if err != nil {
		return Quota{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceQuotaDefinitionRequest,
		URIParams:   Params{"space_quota_guid": quota.GUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Quota{}, nil, err
	}

	var updatedObj Quota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}
