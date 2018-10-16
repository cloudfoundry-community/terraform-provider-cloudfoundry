package cfapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	cferrors "code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/net"
)

// SegmentManager -
type SegmentManager struct {
	log         *Logger
	config      coreconfig.Reader
	ccGateway   net.Gateway
	apiEndpoint string
}

// CCSegmentResource -
type CCSegmentResource struct {
	Name string `json:"name"`
	GUID string `json:"guid"`
}

// CCSegmentOrg
type CCSegmentOrg struct {
	GUID string `json:"guid"`
}

// CCSegmentOrgs -
type CCSegmentOrgs struct {
	Orgs []CCSegmentOrg `json:"data"`
}

// CCSegmentResponse -
type CCSegmentPaginatedResponse struct {
	Resources []CCSegmentResource `json:"resources"`
}

// newSegmentManager -
func newSegmentManager(config coreconfig.Reader, ccGateway net.Gateway, logger *Logger) (dm *SegmentManager, err error) {
	dm = &SegmentManager{
		log:         logger,
		config:      config,
		ccGateway:   ccGateway,
		apiEndpoint: config.APIEndpoint(),
	}

	if len(dm.apiEndpoint) == 0 {
		return nil, errors.New("API endpoint missing from config file")
	}

	return dm, nil
}

// ReadSegment -
func (sm *SegmentManager) ReadSegment(segID string) (seg CCSegmentResource, err error) {
	path := fmt.Sprintf("%s/v3/isolation_segments/%s", sm.apiEndpoint, segID)
	if err = sm.ccGateway.GetResource(path, &seg); err != nil {
		return CCSegmentResource{}, err
	}
	return seg, nil
}

// CreateSegment -
func (sm *SegmentManager) CreateSegment(name string) (seg CCSegmentResource, err error) {
	payload := map[string]interface{}{"name": name}
	body, err := json.Marshal(payload)
	if err != nil {
		return CCSegmentResource{}, err
	}

	err = sm.ccGateway.CreateResource(sm.apiEndpoint, "/v3/isolation_segments", bytes.NewReader(body), &seg)
	if err != nil {
		return CCSegmentResource{}, err
	}
	return seg, nil
}

// UpdateSegment -
func (sm *SegmentManager) UpdateSegment(id string, name string) (seg CCSegmentResource, err error) {
	payload := map[string]interface{}{"name": name}
	body, err := json.Marshal(payload)
	if err != nil {
		return CCSegmentResource{}, err
	}
	path := fmt.Sprintf("/v3/isolation_segments/%s", id)
	err = sm.patchResource(sm.apiEndpoint, path, bytes.NewReader(body), &seg)
	if err != nil {
		return CCSegmentResource{}, err
	}
	return seg, nil
}

// This one should belong to gateway.go, but that API is deprecated
func (sm *SegmentManager) patchResource(endpoint, apiURL string, body io.ReadSeeker, optionalResource ...interface{}) error {
	var resource interface{}
	if len(optionalResource) > 0 {
		resource = optionalResource[0]
	}

	request, err := sm.ccGateway.NewRequest("PATCH", endpoint+apiURL, sm.config.AccessToken(), body)
	if err != nil {
		return err
	}

	if resource == nil {
		_, err = sm.ccGateway.PerformRequest(request)
		return err
	}

	// as sync is false - we removed this from the if
	if sm.ccGateway.PollingEnabled {
		_, err = sm.ccGateway.PerformPollingRequestForJSONResponse(endpoint, request, resource, sm.ccGateway.AsyncTimeout())
		return err
	}

	_, err = sm.ccGateway.PerformRequestForJSONResponse(request, resource)
	if err != nil {
		return err
	}

	return nil
}

// FindSegment -
/// TODO : handle pagination properly, here we have at most 1 result, watting for v3 cli bindings
func (sm *SegmentManager) FindSegment(name string) (CCSegmentResource, error) {
	resource := CCSegmentPaginatedResponse{}
	path := fmt.Sprintf("%s/v3/isolation_segments?names=%s", sm.apiEndpoint, name)
	err := sm.ccGateway.GetResource(path, &resource)
	if err != nil {
		return CCSegmentResource{}, err
	}

	if len(resource.Resources) == 0 {
		return CCSegmentResource{}, cferrors.NewModelNotFoundError("CCSegmentResource", name)
	}

	return resource.Resources[0], nil
}

// DeleteSegment -
func (sm *SegmentManager) DeleteSegment(id string) (err error) {
	path := fmt.Sprintf("/v3/isolation_segments/%s", id)
	return sm.ccGateway.DeleteResource(sm.apiEndpoint, path)
}

// SetSegmentOrgs -
func (sm *SegmentManager) SetSegmentOrgs(id string, orgs []interface{}) (err error) {
	payload := CCSegmentOrgs{}
	for _, org := range orgs {
		payload.Orgs = append(payload.Orgs, CCSegmentOrg{org.(string)})
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations", id)
	err = sm.ccGateway.CreateResource(sm.apiEndpoint, path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	return nil
}

// DeleteSegmentOrgs -
func (sm *SegmentManager) DeleteSegmentOrgs(id string, orgs []interface{}) (err error) {
	for _, org := range orgs {
		path := fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations/%s", id, org)
		err = sm.ccGateway.DeleteResourceSynchronously(sm.apiEndpoint, path)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetSegmentOrgs -
func (sm *SegmentManager) GetSegmentOrgs(ID string) (orgs []interface{}, err error) {
	path := fmt.Sprintf("%s/v3/isolation_segments/%s/relationships/organizations", sm.apiEndpoint, ID)
	resource := CCSegmentOrgs{}
	if err := sm.ccGateway.GetResource(path, &resource); err != nil {
		return orgs, err
	}
	for _, org := range resource.Orgs {
		orgs = append(orgs, org.GUID)
	}
	return orgs, nil
}
