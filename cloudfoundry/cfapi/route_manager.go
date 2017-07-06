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

// newRouteManager -
func newRouteManager(config coreconfig.Reader, ccGateway net.Gateway, logger *Logger) (rm *RouteManager, err error) {

	rm = &RouteManager{
		log: logger,

		config:    config,
		ccGateway: ccGateway,

		apiEndpoint: config.APIEndpoint(),

		repo: api.NewCloudControllerRouteRepository(config, ccGateway),
	}

	return
}

// FindRoute -
func (rm *RouteManager) FindRoute(domain string, hostname, path *string, port *int) (route CCRoute, err error) {

	var apiPath string

	if hostname != nil {
		apiPath = "/v2/routes?q=host:" + *hostname
	} else {
		apiPath = "/v2/routes"
	}

	if err = rm.ccGateway.ListPaginatedResources(rm.apiEndpoint,
		apiPath, CCRouteResource{}, func(resource interface{}) bool {

			routeResource := resource.(CCRouteResource)

			if path != nil && path != routeResource.Entity.Path {
				return true
			}
			if port != nil && port != routeResource.Entity.Port {
				return true
			}

			domainResource := CCDomainResource{}
			err = rm.ccGateway.GetResource(fmt.Sprintf("%s/v2/shared_domains/%s",
				rm.apiEndpoint, routeResource.Entity.DomainGUID), &domainResource)

			if domain != domainResource.Entity.Name {
				return true
			}

			route = routeResource.Entity
			route.ID = routeResource.Metadata.GUID

			return false
		}); err != nil {

		return
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
		err = errors.NewModelNotFoundError("Route", name)
	}
	return
}

// ReadRoute -
func (rm *RouteManager) ReadRoute(routeID string) (route CCRoute, err error) {

	resource := CCRouteResource{}
	err = rm.ccGateway.GetResource(
		fmt.Sprintf("%s/v2/routes/%s", rm.apiEndpoint, routeID), &resource)

	route = resource.Entity
	route.ID = resource.Metadata.GUID
	return
}

// CreateRoute -
func (rm *RouteManager) CreateRoute(r CCRoute) (route CCRoute, err error) {

	body, err := json.Marshal(r)
	if err != nil {
		return
	}

	resource := CCRouteResource{}
	if err = rm.ccGateway.CreateResource(rm.apiEndpoint,
		"/v2/routes", bytes.NewReader(body), &resource); err != nil {
		return
	}
	route = resource.Entity
	route.ID = resource.Metadata.GUID
	return
}

// UpdateRoute -
func (rm *RouteManager) UpdateRoute(r CCRoute) (route CCRoute, err error) {

	body, err := json.Marshal(r)
	if err != nil {
		return
	}

	request, err := rm.ccGateway.NewRequest("PUT",
		fmt.Sprintf("%s/v2/routes/%s", rm.apiEndpoint, r.ID),
		rm.config.AccessToken(), bytes.NewReader(body))
	if err != nil {
		return
	}

	resource := CCRouteResource{}
	_, err = rm.ccGateway.PerformRequestForJSONResponse(request, &resource)

	route = resource.Entity
	route.ID = resource.Metadata.GUID
	return
}

// DeleteRoute -
func (rm *RouteManager) DeleteRoute(routeID string) (err error) {
	err = rm.ccGateway.DeleteResource(rm.apiEndpoint, fmt.Sprintf("/v2/routes/%s", routeID))
	return
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
		return
	}

	response := make(map[string]interface{})
	if err = rm.ccGateway.CreateResource(rm.apiEndpoint,
		"/v2/route_mappings", bytes.NewReader(body), &response); err != nil {
		return
	}

	mappingID = response["metadata"].(map[string]interface{})["guid"].(string)
	return
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

	if err = rm.ccGateway.ListPaginatedResources(rm.apiEndpoint,
		fmt.Sprintf("/v2/route_mappings?q=%s:%s", key, id),
		resource, func(resource interface{}) bool {

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

		}); err != nil {

		return
	}
	return
}

// DeleteRouteMapping -
func (rm *RouteManager) DeleteRouteMapping(mappingID string) (err error) {
	err = rm.ccGateway.DeleteResource(rm.apiEndpoint, fmt.Sprintf("/v2/route_mappings/%s", mappingID))
	return
}
