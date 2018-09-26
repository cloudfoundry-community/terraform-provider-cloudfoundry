package cfapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/net"
)

// RouteManager -
type RouteManager struct {
	log *Logger

	config    coreconfig.Reader
	ccGateway net.Gateway

	apiEndpoint string
	repo        api.RouteRepository
}

// CCRoute -
type CCRoute struct {
	ID string

	DomainGUID string  `json:"domain_guid,omitempty"`
	SpaceGUID  string  `json:"space_guid,omitempty"`
	Hostname   *string `json:"host,omitempty"`
	Port       *int    `json:"port,omitempty"`
	Path       *string `json:"path,omitempty"`
}

// CCRouteResource -
type CCRouteResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCRoute            `json:"entity"`
}

// CCRouteMapping -
type CCRouteMapping struct {
	ID string

	AppPort  string `json:"app_port"`
	AppID    string `json:"app_guid"`
	RouteID  string `json:"route_guid"`
	AppURL   string `json:"app_url"`
	RouteURL string `json:"route_url"`
}

// CCRouteMappingResource -
type CCRouteMappingResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCRouteMapping     `json:"entity"`
}

// newRouteManager -
func newRouteManager(config coreconfig.Reader, ccGateway net.Gateway, logger *Logger) (rm *RouteManager, err error) {
	rm = &RouteManager{
		log:         logger,
		config:      config,
		ccGateway:   ccGateway,
		apiEndpoint: config.APIEndpoint(),
		repo:        api.NewCloudControllerRouteRepository(config, ccGateway),
	}

	return rm, nil
}

// FindRoute -
func (rm *RouteManager) FindRoute(
	domain string,
	hostname,
	path *string,
	port *int) (route CCRoute, err error) {

	var apiPath string

	if hostname != nil {
		apiPath = "/v2/routes?q=host:" + *hostname
	} else {
		apiPath = "/v2/routes"
	}

	if err = rm.ccGateway.ListPaginatedResources(rm.apiEndpoint, apiPath, CCRouteResource{},
		func(resource interface{}) bool {
			routeResource := resource.(CCRouteResource)
			if path != nil && path != routeResource.Entity.Path {
				return true
			}
			if port != nil && port != routeResource.Entity.Port {
				return true
			}
			domainResource := CCDomainResource{}
			url := fmt.Sprintf("%s/v2/shared_domains/%s", rm.apiEndpoint, routeResource.Entity.DomainGUID)
			err = rm.ccGateway.GetResource(url, &domainResource)
			if domain != domainResource.Entity.Name {
				return true
			}
			route = routeResource.Entity
			route.ID = routeResource.Metadata.GUID
			return false
		}); err != nil {
		return CCRoute{}, err
	}

	if len(route.ID) == 0 {
		name := domain
		if hostname != nil {
			name = *hostname + "." + name
		}
		if port != nil {
			name = name + ":" + strconv.Itoa(*port)
		}
		if path != nil {
			name = name + *path
		}
		return CCRoute{}, errors.NewModelNotFoundError("Route", name)
	}

	return route, nil
}

// ReadRoute -
func (rm *RouteManager) ReadRoute(routeID string) (route CCRoute, err error) {
	resource := CCRouteResource{}
	path := fmt.Sprintf("%s/v2/routes/%s", rm.apiEndpoint, routeID)
	if err = rm.ccGateway.GetResource(path, &resource); err != nil {
		return CCRoute{}, err
	}
	route = resource.Entity
	route.ID = resource.Metadata.GUID
	return route, nil
}

// CreateRoute -
func (rm *RouteManager) CreateRoute(r CCRoute, randomPort bool) (route CCRoute, err error) {

	body, err := json.Marshal(r)
	if err != nil {
		return CCRoute{}, err
	}

	var path string
	if randomPort {
		path = "/v2/routes?generate_port=true"
	} else {
		path = "/v2/routes"
	}

	resource := CCRouteResource{}
	if err = rm.ccGateway.CreateResource(rm.apiEndpoint, path, bytes.NewReader(body), &resource); err != nil {
		return CCRoute{}, err
	}
	route = resource.Entity
	route.ID = resource.Metadata.GUID
	return route, nil
}

// UpdateRoute -
func (rm *RouteManager) UpdateRoute(r CCRoute) (route CCRoute, err error) {
	body, err := json.Marshal(r)
	if err != nil {
		return CCRoute{}, err
	}

	path := fmt.Sprintf("%s/v2/routes/%s", rm.apiEndpoint, r.ID)
	request, err := rm.ccGateway.NewRequest("PUT", path, rm.config.AccessToken(), bytes.NewReader(body))
	if err != nil {
		return CCRoute{}, err
	}

	resource := CCRouteResource{}
	if _, err = rm.ccGateway.PerformRequestForJSONResponse(request, &resource); err != nil {
		return CCRoute{}, err
	}

	route = resource.Entity
	route.ID = resource.Metadata.GUID
	return route, nil
}

// DeleteRoute -
func (rm *RouteManager) DeleteRoute(routeID string) (err error) {
	return rm.ccGateway.DeleteResource(rm.apiEndpoint, fmt.Sprintf("/v2/routes/%s", routeID))
}

// CreateRouteMapping -
func (rm *RouteManager) CreateRouteMapping(routeID, appID string, port *int) (mappingID string, err error) {
	request := map[string]interface{}{
		"app_guid":   appID,
		"route_guid": routeID,
	}
	if port != nil {
		request["app_port"] = *port
	}
	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	response := make(map[string]interface{})
	if err = rm.ccGateway.CreateResource(rm.apiEndpoint, "/v2/route_mappings", bytes.NewReader(body), &response); err != nil {
		return "", err
	}

	mappingID = response["metadata"].(map[string]interface{})["guid"].(string)
	return mappingID, nil
}

// ReadRouteMapping -
func (rm *RouteManager) ReadRouteMapping(mappingID string) (CCRouteMapping, error) {
	resource := CCRouteMappingResource{}
	path := fmt.Sprintf("/v2/route_mappings/%s", mappingID)
	if err := rm.ccGateway.GetResource(path, &resource); err != nil {
		return CCRouteMapping{}, err
	}
	routeMapping := resource.Entity
	routeMapping.ID = resource.Metadata.GUID
	return routeMapping, nil

}

// ReadRouteMappingsByRoute -
func (rm *RouteManager) ReadRouteMappingsByRoute(routeID string) ([]map[string]interface{}, error) {
	return rm.readRouteMappings(routeID, "route_guid")
}

// ReadRouteMappingsByApp -
func (rm *RouteManager) ReadRouteMappingsByApp(appID string) (mappings []map[string]interface{}, err error) {
	return rm.readRouteMappings(appID, "app_guid")
}

// readRouteMappings -
func (rm *RouteManager) readRouteMappings(id, key string) (mappings []map[string]interface{}, err error) {

	resource := make(map[string]interface{})
	path := fmt.Sprintf("/v2/route_mappings?q=%s:%s", key, id)
	err = rm.ccGateway.ListPaginatedResources(rm.apiEndpoint, path, resource, func(resource interface{}) bool {
		routeResource := resource.(map[string]interface{})
		mapping := make(map[string]interface{})
		mapping["mapping_id"] = routeResource["metadata"].(map[string]interface{})["guid"].(string)
		switch key {
		case "route_guid":
			mapping["app"] = routeResource["entity"].(map[string]interface{})["app_guid"].(string)
		case "app_guid":
			mapping["route"] = routeResource["entity"].(map[string]interface{})["route_guid"].(string)
		default:
			mapping["app"] = routeResource["entity"].(map[string]interface{})["app_guid"].(string)
			mapping["route"] = routeResource["entity"].(map[string]interface{})["route_guid"].(string)
		}
		if v, ok := routeResource["entity"].(map[string]interface{})["app_port"]; ok {
			mapping["port"] = int(v.(float64))
		}
		mappings = append(mappings, mapping)
		return true
	})
	return mappings, err
}

// DeleteRouteMapping -
func (rm *RouteManager) DeleteRouteMapping(mappingID string) (err error) {
	return rm.ccGateway.DeleteResource(rm.apiEndpoint, fmt.Sprintf("/v2/route_mappings/%s", mappingID))
}
