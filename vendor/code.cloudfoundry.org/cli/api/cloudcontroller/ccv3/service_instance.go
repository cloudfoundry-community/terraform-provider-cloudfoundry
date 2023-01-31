package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/lookuptable"
)

type SpaceWithOrganization struct {
	SpaceGUID        string
	SpaceName        string
	OrganizationName string
}

// GetServiceInstances lists service instances with optional filters.
func (client *Client) GetServiceInstances(query ...Query) ([]resources.ServiceInstance, IncludedResources, Warnings, error) {
	var result []resources.ServiceInstance

	included, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServiceInstancesRequest,
		Query:        query,
		ResponseBody: resources.ServiceInstance{},
		AppendToList: func(item interface{}) error {
			result = append(result, item.(resources.ServiceInstance))
			return nil
		},
	})

	return result, included, warnings, err
}

func (client *Client) GetServiceInstanceByNameAndSpace(name, spaceGUID string, query ...Query) (resources.ServiceInstance, IncludedResources, Warnings, error) {
	query = append(query,
		Query{
			Key:    NameFilter,
			Values: []string{name},
		},
		Query{
			Key:    SpaceGUIDFilter,
			Values: []string{spaceGUID},
		},
	)

	instances, included, warnings, err := client.GetServiceInstances(query...)

	if err != nil {
		return resources.ServiceInstance{}, IncludedResources{}, warnings, err
	}

	if len(instances) == 0 {
		return resources.ServiceInstance{},
			IncludedResources{},
			warnings,
			ccerror.ServiceInstanceNotFoundError{
				Name:      name,
				SpaceGUID: spaceGUID,
			}
	}

	return instances[0], included, warnings, nil
}

func (client *Client) GetServiceInstanceParameters(serviceInstanceGUID string) (parameters types.JSONObject, warnings Warnings, err error) {
	_, warnings, err = client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceInstanceParametersRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		ResponseBody: &parameters,
	})

	return
}

// GetUserProvidedServiceInstanceCredentails get credentials for a user-provided service instance only
// Does not return async job url
func (client *Client) GetUserProvidedServiceInstanceCredentails(serviceInstanceGUID string) (types.JSONObject, Warnings, error) {
	var creds types.JSONObject
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceInstanceCredentialsRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		ResponseBody: &creds,
	})

	return creds, warnings, err
}

func (client *Client) CreateServiceInstance(serviceInstance resources.ServiceInstance) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{
		RequestName: internal.PostServiceInstanceRequest,
		RequestBody: serviceInstance,
	})
}

// CreateUserProvidedServiceInstance return the service instance in response body because create requests for User-provided service instances don't return jobURLs
func (client *Client) CreateUserProvidedServiceInstance(serviceInstance resources.ServiceInstance) (resources.ServiceInstance, Warnings, error) {
	var responseBody resources.ServiceInstance
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostServiceInstanceRequest,
		RequestBody:  serviceInstance,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateUserProvidedServiceInstance return the service instance in response body instead of jobURLs
func (client *Client) UpdateUserProvidedServiceInstance(serviceInstanceGUID string, serviceInstanceUpdates resources.ServiceInstance) (resources.ServiceInstance, Warnings, error) {
	var responseBody resources.ServiceInstance

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchServiceInstanceRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		RequestBody:  serviceInstanceUpdates,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) UpdateServiceInstance(serviceInstanceGUID string, serviceInstanceUpdates resources.ServiceInstance) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{
		RequestName: internal.PatchServiceInstanceRequest,
		URIParams:   internal.Params{"service_instance_guid": serviceInstanceGUID},
		RequestBody: serviceInstanceUpdates,
	})
}

func (client *Client) DeleteServiceInstance(serviceInstanceGUID string, query ...Query) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{
		RequestName: internal.DeleteServiceInstanceRequest,
		URIParams:   internal.Params{"service_instance_guid": serviceInstanceGUID},
		Query:       query,
	})
}

// GetServiceInstanceSharedSpaces will fetch relationships between
// a service instance and the shared-to spaces for that service.
func (client *Client) GetServiceInstanceSharedSpaces(serviceInstanceGUID string) ([]SpaceWithOrganization, Warnings, error) {
	var responseBody resources.SharedToSpacesListWrapper
	query := []Query{
		{
			Key:    FieldsSpace,
			Values: []string{"guid", "name", "relationships.organization"},
		},
		{
			Key:    FieldsSpaceOrganization,
			Values: []string{"guid", "name"},
		},
	}
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceInstanceRelationshipsSharedSpacesRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		Query:        query,
		ResponseBody: &responseBody,
	})
	return mapRelationshipsToSpaces(responseBody), warnings, err
}

func (client *Client) GetServiceInstanceUsageSummary(serviceInstanceGUID string) ([]resources.ServiceInstanceUsageSummary, Warnings, error) {
	var result resources.ServiceInstanceUsageSummaryList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceInstanceSharedSpacesUsageSummaryRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		ResponseBody: &result,
	})
	return result.UsageSummary, warnings, err
}

func mapRelationshipsToSpaces(sharedToSpaces resources.SharedToSpacesListWrapper) []SpaceWithOrganization {
	var spacesToReturn []SpaceWithOrganization

	guidToOrgNameLookup := lookuptable.NameFromGUID(sharedToSpaces.Organizations)

	for _, s := range sharedToSpaces.Spaces {
		org := s.Relationships[constant.RelationshipTypeOrganization]
		space := SpaceWithOrganization{SpaceGUID: s.GUID, SpaceName: s.Name, OrganizationName: guidToOrgNameLookup[org.GUID]}
		spacesToReturn = append(spacesToReturn, space)
	}

	return spacesToReturn
}
