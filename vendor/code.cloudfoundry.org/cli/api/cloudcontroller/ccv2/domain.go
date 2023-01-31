package ccv2

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Domain represents a Cloud Controller Domain.
type Domain struct {
	// GUID is the unique domain identifier.
	GUID string

	// Internal indicates whether the domain is an internal domain.
	Internal bool

	// Name is the name given to the domain.
	Name string

	// RouterGroupGUID is the unique identier of the router group this domain is
	// assigned to.
	RouterGroupGUID string

	// RouterGroupType is the type of router group this domain is assigned to. It
	// can be of type `tcp` or `http`.
	RouterGroupType constant.RouterGroupType

	// DomainType is the access type of the domain. It can be either a domain
	// private to a single org or it can be a domain shared to all orgs.
	Type constant.DomainType

	// Organization owning the domain
	OwningOrganizationGUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Domain response.
func (domain *Domain) UnmarshalJSON(data []byte) error {
	var ccDomain struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name                   string `json:"name"`
			RouterGroupGUID        string `json:"router_group_guid"`
			RouterGroupType        string `json:"router_group_type"`
			Internal               bool   `json:"internal"`
			OwningOrganizationGUID string `json:"owning_organization_guid"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccDomain)
	if err != nil {
		return err
	}

	domain.GUID = ccDomain.Metadata.GUID
	domain.Name = ccDomain.Entity.Name
	domain.RouterGroupGUID = ccDomain.Entity.RouterGroupGUID
	domain.RouterGroupType = constant.RouterGroupType(ccDomain.Entity.RouterGroupType)
	domain.Internal = ccDomain.Entity.Internal
	domain.OwningOrganizationGUID = ccDomain.Entity.OwningOrganizationGUID
	return nil
}

type createSharedDomainBody struct {
	Name            string `json:"name"`
	RouterGroupGUID string `json:"router_group_guid,omitempty"`
	Internal        bool   `json:"internal"`
}

type createPrivateDomainBody struct {
	Name                   string `json:"name"`
	OwningOrganizationGuid string `json:"owning_organization_guid"`
}

func (client *Client) CreateSharedDomain(domainName string, routerGroupdGUID string, isInternal bool) (Domain, Warnings, error) {
	body := createSharedDomainBody{
		Name:            domainName,
		RouterGroupGUID: routerGroupdGUID,
		Internal:        isInternal,
	}
	bodyBytes, err := json.Marshal(body)
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSharedDomainRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Domain{}, nil, err
	}
	var updatedObj Domain
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}

// GetOrganizationPrivateDomains returns the private domains associated with an organization.
func (client *Client) GetOrganizationPrivateDomains(orgGUID string, filters ...Filter) ([]Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationPrivateDomainsRequest,
		Query:       ConvertFilterParameters(filters),
		URIParams:   map[string]string{"organization_guid": orgGUID},
	})
	if err != nil {
		return []Domain{}, nil, err
	}

	fullDomainsList := []Domain{}
	warnings, err := client.paginate(request, Domain{}, func(item interface{}) error {
		if domain, ok := item.(Domain); ok {
			domain.Type = constant.PrivateDomain
			fullDomainsList = append(fullDomainsList, domain)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Domain{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullDomainsList, warnings, err
}

// GetPrivateDomain returns the Private Domain associated with the provided
// Domain GUID.
func (client *Client) GetPrivateDomain(domainGUID string) (Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetPrivateDomainRequest,
		URIParams:   map[string]string{"private_domain_guid": domainGUID},
	})
	if err != nil {
		return Domain{}, nil, err
	}

	var domain Domain
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &domain,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return Domain{}, response.Warnings, err
	}

	domain.Type = constant.PrivateDomain
	return domain, response.Warnings, nil
}

// GetPrivateDomains returns the private domains this client has access to.
func (client *Client) GetPrivateDomains(filters ...Filter) ([]Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetPrivateDomainsRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return []Domain{}, nil, err
	}

	fullDomainsList := []Domain{}
	warnings, err := client.paginate(request, Domain{}, func(item interface{}) error {
		if domain, ok := item.(Domain); ok {
			domain.Type = constant.PrivateDomain
			fullDomainsList = append(fullDomainsList, domain)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Domain{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullDomainsList, warnings, err
}

// DeletePrivateDomain delete a private domain
func (client *Client) DeletePrivateDomain(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeletePrivateDomainRequest,
		URIParams:   map[string]string{"private_domain_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// GetSharedDomain returns the Shared Domain associated with the provided
// Domain GUID.
func (client *Client) GetSharedDomain(domainGUID string) (Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSharedDomainRequest,
		URIParams:   map[string]string{"shared_domain_guid": domainGUID},
	})
	if err != nil {
		return Domain{}, nil, err
	}

	var domain Domain
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &domain,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return Domain{}, response.Warnings, err
	}

	domain.Type = constant.SharedDomain
	return domain, response.Warnings, nil
}

// DeleteSharedDomain delete a shared domain
func (client *Client) DeleteSharedDomain(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSharedDomainRequest,
		URIParams:   map[string]string{"shared_domain_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// GetSharedDomains returns the global shared domains.
func (client *Client) GetSharedDomains(filters ...Filter) ([]Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSharedDomainsRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return []Domain{}, nil, err
	}

	fullDomainsList := []Domain{}
	warnings, err := client.paginate(request, Domain{}, func(item interface{}) error {
		if domain, ok := item.(Domain); ok {
			domain.Type = constant.SharedDomain
			fullDomainsList = append(fullDomainsList, domain)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Domain{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullDomainsList, warnings, err
}

// DeleteOrganizationPrivateDomain delete a organization private domain
func (client *Client) DeleteOrganizationPrivateDomain(organizationGuid, privateDomainGuid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationPrivateDomainRequest,
		URIParams: Params{
			"organization_guid":   organizationGuid,
			"private_domain_guid": privateDomainGuid,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// CreateDomain creates a cloud controller domain in with the given settings.
func (client *Client) CreatePrivateDomain(domainName string, organizationGUID string) (Domain, Warnings, error) {
	body := createPrivateDomainBody{
		Name:                   domainName,
		OwningOrganizationGuid: organizationGUID,
	}
	bodyBytes, err := json.Marshal(body)
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostPrivateDomainRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Domain{}, nil, err
	}
	var updatedObj Domain
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedObj,
	}

	err = client.connection.Make(request, &response)
	return updatedObj, response.Warnings, err
}

// CreateDomain creates a cloud controller domain in with the given settings.
func (client *Client) SetOrganizationPrivateDomain(domainGUID string, organizationGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationPrivateDomainRequest,
		URIParams: Params{
			"organization_guid":   organizationGUID,
			"private_domain_guid": domainGUID,
		},
	})
	if err != nil {
		return nil, err
	}
	var response cloudcontroller.Response

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}
