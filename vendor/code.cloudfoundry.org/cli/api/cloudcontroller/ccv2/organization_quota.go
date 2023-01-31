package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"encoding/json"
)

// GetOrganizationQuota returns an Organization Quota associated with the
// provided GUID.
func (client *Client) GetOrganizationQuota(guid string) (Quota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationQuotaDefinitionRequest,
		URIParams:   Params{"organization_quota_guid": guid},
	})
	if err != nil {
		return Quota{}, nil, err
	}

	var orgQuota Quota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &orgQuota,
	}

	err = client.connection.Make(request, &response)
	return orgQuota, response.Warnings, err
}

// GetOrganizationQuotas returns an Organization Quota list associated with the
// provided filters.
func (client *Client) GetOrganizationQuotas(filters ...Filter) ([]Quota, Warnings, error) {
	allQueries := ConvertFilterParameters(filters)
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationQuotaDefinitionsRequest,
		Query:       allQueries,
	})

	if err != nil {
		return []Quota{}, nil, err
	}

	var fullOrgQuotasList []Quota

	warnings, err := client.paginate(request, Quota{}, func(item interface{}) error {
		if org, ok := item.(Quota); ok {
			fullOrgQuotasList = append(fullOrgQuotasList, org)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Quota{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullOrgQuotasList, warnings, err
}

// UpdateQuota updates the quota with the given GUID.
func (client *Client) UpdateOrganizationQuota(quota Quota) (Quota, Warnings, error) {
	body, err := json.Marshal(quota)
	if err != nil {
		return Quota{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationQuotaDefinitionRequest,
		URIParams:   Params{"organization_quota_guid": quota.GUID},
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
func (client *Client) DeleteOrganizationQuota(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationQuotaDefinitionRequest,
		URIParams: Params{
			"organization_quota_guid": guid,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// CreateQuota creates a cloud controller quota in with the given settings.
func (client *Client) CreateOrganizationQuota(quota Quota) (Quota, Warnings, error) {
	body, err := json.Marshal(quota)
	if err != nil {
		return Quota{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostOrganizationQuotaDefinitionsRequest,
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
