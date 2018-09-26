package cfapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// ServiceManager -
type ServiceManager struct {
	log *Logger

	config    coreconfig.Reader
	ccGateway net.Gateway

	apiEndpoint string

	repo   api.ServiceRepository
	sbRepo api.ServiceBrokerRepository
}

// CCService -
type CCService struct {
	ID string

	ServiceBrokerGUID string `json:"service_broker_guid,omitempty"`

	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`

	Active         bool `json:"active,omitempty"`
	Bindable       bool `json:"bindable,omitempty"`
	PlanUpdateable bool `json:"plan_updateable,omitempty"`

	Extra string `json:"extra,omitempty"`

	Tags     []string `json:"tags,omitempty"`
	Requires []string `json:"requires,omitempty"`

	ServicePlans []CCServicePlan
}

// CCServiceResource -
type CCServiceResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCService          `json:"entity"`
}

// CCServiceResourceList -
type CCServiceResourceList struct {
	Resources []CCServiceResource `json:"resources"`
}

// CCServicePlan -
type CCServicePlan struct {
	ID string

	Name        string `json:"name"`
	Description string `json:"description"`

	Free   bool `json:"free"`
	Public bool `json:"public"`
	Active bool `json:"active"`
}

// CCServicePlanResource -
type CCServicePlanResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCServicePlan      `json:"entity"`
}

// CCServicePlanResourceList -
type CCServicePlanResourceList struct {
	Resources []CCServicePlanResource `json:"resources"`
}

// CCServiceBroker -
type CCServiceBroker struct {
	Name         string `json:"name,omitempty"`
	BrokerURL    string `json:"broker_url,omitempty"`
	AuthUserName string `json:"auth_username,omitempty"`
	AuthPassword string `json:"auth_password,omitempty"`
	SpaceGUID    string `json:"space_guid,omitempty"`
}

// CCServiceBrokerResource -
type CCServiceBrokerResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCServiceBroker    `json:"entity"`
}

// CCServiceInstance -
type CCServiceInstance struct {
	Name            string                 `json:"name"`
	SpaceGUID       string                 `json:"space_guid"`
	ServicePlanGUID string                 `json:"service_plan_guid"`
	Tags            []string               `json:"tags,omitempty"`
	LastOperation   map[string]interface{} `json:"last_operation"`
}

// CCServiceInstanceResource -
type CCServiceInstanceResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCServiceInstance  `json:"entity"`
}

// CCServiceInstanceUpdateRequest -
type CCServiceInstanceUpdateRequest struct {
	Name            string                 `json:"name"`
	ServicePlanGUID string                 `json:"service_plan_guid"`
	Params          map[string]interface{} `json:"parameters,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
}

// CCUserProvidedService -
type CCUserProvidedService struct {
	Name            string                 `json:"name"`
	SpaceGUID       string                 `json:"space_guid"`
	SyslogDrainURL  string                 `json:"syslog_drain_url,omitempty"`
	RouteServiceURL string                 `json:"route_service_url,omitempty"`
	Credentials     map[string]interface{} `json:"credentials,omitempty"`
}

// CCUserProvidedServiceResource -
type CCUserProvidedServiceResource struct {
	Metadata resources.Metadata    `json:"metadata"`
	Entity   CCUserProvidedService `json:"entity"`
}

// CCUserProvidedServiceUpdateRequest -
type CCUserProvidedServiceUpdateRequest struct {
	Name            string                 `json:"name"`
	ServicePlanGUID string                 `json:"service_plan_guid"`
	SyslogDrainURL  string                 `json:"syslog_drain_url,omitempty"`
	RouteServiceURL string                 `json:"route_service_url,omitempty"`
	Credentials     map[string]interface{} `json:"credentials,omitempty"`
}

// CCServiceKey -
type CCServiceKey struct {
	ID string

	Name        string                 `json:"name"`
	ServiceGUID string                 `json:"service_instance_guid"`
	Credentials map[string]interface{} `json:"credentials"`
}

// CCServiceKeyResource -
type CCServiceKeyResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCServiceKey       `json:"entity"`
}

// CCServiceInstanceRoute -
type CCServiceInstanceRoute struct {
	Host string `json:"host"`
}

// CCServiceInstanceRouteResource -
type CCServiceInstanceRouteResource struct {
	Metadata resources.Metadata     `json:"metadata"`
	Entity   CCServiceInstanceRoute `json:"entity"`
}

// NewServiceManager -
func newServiceManager(config coreconfig.Reader, ccGateway net.Gateway, logger *Logger) (sm *ServiceManager, err error) {
	return &ServiceManager{
		log:         logger,
		config:      config,
		ccGateway:   ccGateway,
		apiEndpoint: config.APIEndpoint(),
		repo:        api.NewCloudControllerServiceRepository(config, ccGateway),
		sbRepo:      api.NewCloudControllerServiceBrokerRepository(config, ccGateway),
	}, nil
}

// ReadServiceInfo -
func (sm *ServiceManager) ReadServiceInfo(serviceBrokerID string) (services []CCService, err error) {
	path := fmt.Sprintf("/v2/services?q=service_broker_guid:%s", serviceBrokerID)
	err = sm.ccGateway.ListPaginatedResources(sm.apiEndpoint, path, CCServiceResource{}, func(resource interface{}) bool {
		sr := resource.(CCServiceResource)
		service := sr.Entity
		service.ID = sr.Metadata.GUID

		if err = sm.ccGateway.ListPaginatedResources(sm.apiEndpoint,
			fmt.Sprintf("/v2/services/%s/service_plans", service.ID),
			CCServicePlanResource{}, func(resource interface{}) bool {

				spr := resource.(CCServicePlanResource)
				servicePlan := spr.Entity
				servicePlan.ID = spr.Metadata.GUID

				service.ServicePlans = append(service.ServicePlans, servicePlan)
				return true

			}); err != nil {

			sm.log.DebugMessage("WARNING! Unable to retrieve service plans for service '%s': %s", service.ID, err.Error())
			err = nil
		}

		services = append(services, service)
		return true

	})
	return services, err
}

// CreateServiceBroker -
func (sm *ServiceManager) CreateServiceBroker(name, brokerURL, authUserName, authPassword, spaceGUID string) (id string, err error) {
	path := "/v2/service_brokers"
	request := CCServiceBroker{
		Name:         name,
		BrokerURL:    brokerURL,
		AuthUserName: authUserName,
		AuthPassword: authPassword,
	}
	if len(spaceGUID) > 0 {
		request.SpaceGUID = spaceGUID
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	resource := CCServiceBrokerResource{}
	if err = sm.ccGateway.CreateResource(sm.apiEndpoint, path, bytes.NewReader(body), &resource); err != nil {
		return "", err
	}

	id = resource.Metadata.GUID
	return id, nil
}

// UpdateServiceBroker -
func (sm *ServiceManager) UpdateServiceBroker(
	serviceBrokerID,
	name,
	brokerURL,
	authUserName,
	authPassword,
	spaceGUID string) (serviceBroker CCServiceBroker, err error) {

	path := fmt.Sprintf("/v2/service_brokers/%s", serviceBrokerID)
	request := CCServiceBroker{
		Name:         name,
		BrokerURL:    brokerURL,
		AuthUserName: authUserName,
		AuthPassword: authPassword,
	}
	if len(spaceGUID) > 0 {
		request.SpaceGUID = spaceGUID
	}

	body, err := json.Marshal(request)
	if err != nil {
		return serviceBroker, err
	}

	resource := CCServiceBrokerResource{}
	if err = sm.ccGateway.UpdateResource(sm.apiEndpoint, path, bytes.NewReader(body), &resource); err != nil {
		return serviceBroker, err
	}

	serviceBroker = resource.Entity
	return serviceBroker, nil
}

// ReadServiceBroker -
func (sm *ServiceManager) ReadServiceBroker(serviceBrokerID string) (serviceBroker CCServiceBroker, err error) {

	url := fmt.Sprintf("%s/v2/service_brokers/%s", sm.apiEndpoint, serviceBrokerID)

	resource := CCServiceBrokerResource{}
	err = sm.ccGateway.GetResource(url, &resource)
	if err != nil {
		return serviceBroker, err
	}

	serviceBroker = resource.Entity
	return serviceBroker, nil
}

// DeleteServiceBroker -
func (sm *ServiceManager) DeleteServiceBroker(serviceBrokerID string) (err error) {

	err = sm.ccGateway.DeleteResource(sm.apiEndpoint, fmt.Sprintf("/v2/service_brokers/%s", serviceBrokerID))
	return err
}

// ForceDeleteServiceBroker -
func (sm *ServiceManager) ForceDeleteServiceBroker(serviceBrokerID string) (err error) {

	services, err := sm.ReadServiceInfo(serviceBrokerID)
	if err != nil {
		return err
	}

	for _, s := range services {
		for _, sp := range s.ServicePlans {
			if err = sm.ccGateway.ListPaginatedResources(
				sm.apiEndpoint,
				fmt.Sprintf("/v2/service_instances?q=service_plan_guid:%s", sp.ID),
				CCServiceInstanceResource{},
				func(resource interface{}) bool {
					sir := resource.(CCServiceInstanceResource)
					if err = sm.ccGateway.DeleteResource(sm.apiEndpoint,
						fmt.Sprintf("/v2/service_instances/%s?purge=true", sir.Metadata.GUID)); err != nil {
						sm.log.DebugMessage("WARNING! Unable to delete service instance '%s': %s", sir.Metadata.GUID, err.Error())
						err = nil
					}
					return true
				}); err != nil {
				sm.log.DebugMessage("WARNING! Unable to retrieve service instances for service '%s': %s", sp.ID, err.Error())
			}
		}
	}

	return sm.DeleteServiceBroker(serviceBrokerID)
}

// GetServiceBrokerID -
func (sm *ServiceManager) GetServiceBrokerID(name string) (id string, err error) {
	sb, err := sm.sbRepo.FindByName(name)
	if err != nil {
		return "", err
	}
	id = sb.GUID
	return id, nil
}

// CreateServicePlanAccess -
func (sm *ServiceManager) CreateServicePlanAccess(servicePlanGUID, orgGUID string) (servicePlanAccessGUID string, err error) {
	path := "/v2/service_plan_visibilities"
	request := map[string]string{
		"service_plan_guid": servicePlanGUID,
		"organization_guid": orgGUID,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	response := make(map[string]interface{})
	err = sm.ccGateway.CreateResource(sm.apiEndpoint, path, bytes.NewReader(body), &response)
	if err != nil {
		return "", err
	}
	servicePlanAccessGUID = response["metadata"].(map[string]interface{})["guid"].(string)
	return servicePlanAccessGUID, nil
}

// UpdateServicePlanAccess -
func (sm *ServiceManager) UpdateServicePlanAccess(
	servicePlanAccessGUID,
	servicePlanGUID,
	orgGUID string) (err error) {

	path := fmt.Sprintf("/v2/service_plan_visibilities/%s", servicePlanAccessGUID)
	request := map[string]string{
		"service_plan_guid": servicePlanGUID,
		"organization_guid": orgGUID,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	response := make(map[string]interface{})
	err = sm.ccGateway.UpdateResource(sm.apiEndpoint, path, bytes.NewReader(body), &response)
	return err
}

// ReadServicePlanAccess -
func (sm *ServiceManager) ReadServicePlanAccess(servicePlanAccessGUID string) (planGUID, orgGUID string, err error) {

	url := fmt.Sprintf("%s/v2/service_plan_visibilities/%s", sm.apiEndpoint, servicePlanAccessGUID)

	response := make(map[string]interface{})
	err = sm.ccGateway.GetResource(url, &response)
	if err != nil {
		return "", "", err
	}

	if entity, ok := response["entity"]; ok {
		planGUID = entity.(map[string]interface{})["service_plan_guid"].(string)
		orgGUID = entity.(map[string]interface{})["organization_guid"].(string)
	} else {
		return "", "", errors.NewModelNotFoundError("service plan access", servicePlanAccessGUID)
	}

	return planGUID, orgGUID, nil
}

// DeleteServicePlanAccess -
func (sm *ServiceManager) DeleteServicePlanAccess(servicePlanAccessGUID string) (err error) {
	err = sm.ccGateway.DeleteResource(sm.apiEndpoint, fmt.Sprintf("/v2/service_plan_visibilities/%s", servicePlanAccessGUID))
	return err
}

// UpdateServicePlanVisibility -
func (sm *ServiceManager) UpdateServicePlanVisibility(planID string, state bool) (err error) {
	path := fmt.Sprintf("/v2/service_plans/%s", planID)
	request := map[string]bool{
		"public": state,
	}
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	ups := CCServicePlanResource{}
	err = sm.ccGateway.UpdateResource(sm.apiEndpoint, path, bytes.NewReader(jsonBytes), &ups)
	return err
}

// ReadServicePlan -
func (sm *ServiceManager) ReadServicePlan(planID string) (CCServicePlan, error) {
	res := CCServicePlanResource{}
	url := fmt.Sprintf("%s/v2/service_plans/%s", sm.apiEndpoint, planID)
	err := sm.ccGateway.GetResource(url, &res)
	return res.Entity, err
}

// CreateServiceInstance -
func (sm *ServiceManager) CreateServiceInstance(
	name,
	servicePlanID,
	spaceID string,
	params map[string]interface{},
	tags []string) (id string, err error) {

	path := "/v2/service_instances?accepts_incomplete=true"
	request := models.ServiceInstanceCreateRequest{
		Name:      name,
		PlanGUID:  servicePlanID,
		SpaceGUID: spaceID,
		Params:    params,
		Tags:      tags,
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	resource := CCServiceInstanceResource{}
	if err = sm.ccGateway.CreateResource(sm.apiEndpoint, path, bytes.NewReader(jsonBytes), &resource); err != nil {
		return "", err
	}

	id = resource.Metadata.GUID
	return id, nil
}

// UpdateServiceInstance -
func (sm *ServiceManager) UpdateServiceInstance(
	serviceInstanceID,
	name,
	servicePlanID string,
	params map[string]interface{},
	tags []string) (serviceInstance CCServiceInstance, err error) {

	path := fmt.Sprintf("/v2/service_instances/%s?accepts_incomplete=true", serviceInstanceID)
	request := CCServiceInstanceUpdateRequest{
		Name:            name,
		ServicePlanGUID: servicePlanID,
		Params:          params,
		Tags:            tags,
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return CCServiceInstance{}, err
	}

	resource := CCServiceInstanceResource{}
	if err = sm.ccGateway.UpdateResource(sm.apiEndpoint, path, bytes.NewReader(jsonBytes), &resource); err != nil {
		return CCServiceInstance{}, err
	}

	serviceInstance = resource.Entity
	return serviceInstance, nil
}

// ReadServiceInstance -
func (sm *ServiceManager) ReadServiceInstance(serviceInstanceID string) (serviceInstance CCServiceInstance, err error) {
	path := fmt.Sprintf("%s/v2/service_instances/%s", sm.apiEndpoint, serviceInstanceID)
	resource := CCServiceInstanceResource{}
	err = sm.ccGateway.GetResource(path, &resource)
	if err != nil {
		return CCServiceInstance{}, err
	}
	serviceInstance = resource.Entity
	return serviceInstance, nil
}

// FindServiceInstance -
func (sm *ServiceManager) FindServiceInstance(name string, spaceID string) (serviceInstance CCServiceInstance, err error) {

	path := fmt.Sprintf("/v2/spaces/%s/service_instances?return_user_provided_service_instances=true&q=%s&inline-relations-depth=1",
		spaceID, url.QueryEscape("name:"+name))

	var found bool

	apiErr := sm.ccGateway.ListPaginatedResources(
		sm.apiEndpoint,
		path,
		CCServiceInstanceResource{},
		func(resource interface{}) bool {
			if sp, ok := resource.(CCServiceInstanceResource); ok {
				serviceInstance = sp.Entity // there should 1 or 0 instances in the space with that name
				found = true
				return false
			}
			return true
		})

	if apiErr != nil {
		switch apiErr.(type) {
		case *errors.HTTPNotFoundError:
			err = errors.NewModelNotFoundError("Space", spaceID)
		default:
			err = apiErr
		}
	} else {
		if !found {
			err = errors.NewModelNotFoundError("ServiceInstance", name)
		}
	}

	return serviceInstance, err
}

// DeleteServiceInstance -
func (sm *ServiceManager) DeleteServiceInstance(serviceInstanceID string, recursive bool) (err error) {

	if !recursive {
		err = sm.ccGateway.DeleteResource(sm.apiEndpoint, fmt.Sprintf("/v2/service_instances/%s?accepts_incomplete=true", serviceInstanceID))
		return err
	}

	err = sm.ccGateway.DeleteResource(sm.apiEndpoint, fmt.Sprintf("/v2/service_instances/%s?recursive=true&accepts_incomplete=true", serviceInstanceID))
	return err
}

// CreateUserProvidedService -
func (sm *ServiceManager) CreateUserProvidedService(
	name string,
	spaceID string,
	credentials map[string]interface{},
	syslogDrainURL string,
	routeServiceURL string) (id string, err error) {

	path := "/v2/user_provided_service_instances"
	request := models.UserProvidedService{
		Name:            name,
		SpaceGUID:       spaceID,
		Credentials:     credentials,
		SysLogDrainURL:  syslogDrainURL,
		RouteServiceURL: routeServiceURL,
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	ups := CCUserProvidedServiceResource{}
	if err = sm.ccGateway.CreateResource(sm.apiEndpoint, path, bytes.NewReader(jsonBytes), &ups); err != nil {
		return "", err
	}

	id = ups.Metadata.GUID
	return id, nil
}

// ReadUserProvidedService -
func (sm *ServiceManager) ReadUserProvidedService(serviceInstanceID string) (ups CCUserProvidedService, err error) {
	path := fmt.Sprintf("%s/v2/user_provided_service_instances/%s", sm.apiEndpoint, serviceInstanceID)
	resource := CCUserProvidedServiceResource{}
	err = sm.ccGateway.GetResource(path, &resource)
	if err != nil {
		return CCUserProvidedService{}, err
	}

	ups = resource.Entity
	return ups, nil
}

// UpdateUserProvidedService -
func (sm *ServiceManager) UpdateUserProvidedService(
	serviceInstanceID string,
	name string,
	credentials map[string]interface{},
	syslogDrainURL string,
	routeServiceURL string) (ups CCUserProvidedService, err error) {

	path := fmt.Sprintf("/v2/user_provided_service_instances/%s", serviceInstanceID)
	request := CCUserProvidedServiceUpdateRequest{
		Name:            name,
		Credentials:     credentials,
		SyslogDrainURL:  syslogDrainURL,
		RouteServiceURL: routeServiceURL,
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return CCUserProvidedService{}, err
	}

	ups = CCUserProvidedService{}
	if err = sm.ccGateway.UpdateResource(sm.apiEndpoint, path, bytes.NewReader(jsonBytes), &ups); err != nil {
		return CCUserProvidedService{}, err
	}

	return ups, nil
}

// DeleteUserProvidedService -
func (sm *ServiceManager) DeleteUserProvidedService(serviceInstanceID string) (err error) {

	err = sm.ccGateway.DeleteResource(sm.apiEndpoint, fmt.Sprintf("/v2/user_provided_service_instances/%s", serviceInstanceID))
	return err
}

// CreateServiceKey -
func (sm *ServiceManager) CreateServiceKey(name, serviceID string, params map[string]interface{}) (serviceKey CCServiceKey, err error) {

	body, err := json.Marshal(map[string]interface{}{
		"name":                  name,
		"service_instance_guid": serviceID,
		"parameters":            params,
	})
	if err != nil {
		return CCServiceKey{}, err
	}

	resource := CCServiceKeyResource{}
	if err = sm.ccGateway.CreateResource(sm.apiEndpoint, "/v2/service_keys", bytes.NewReader(body), &resource); err != nil {
		return CCServiceKey{}, err
	}
	serviceKey = resource.Entity
	serviceKey.ID = resource.Metadata.GUID
	return serviceKey, nil
}

// ReadServiceKey -
func (sm *ServiceManager) ReadServiceKey(serviceKeyID string) (serviceKey CCServiceKey, err error) {

	url := fmt.Sprintf("%s/v2/service_keys/%s", sm.apiEndpoint, serviceKeyID)

	resource := CCServiceKeyResource{}
	err = sm.ccGateway.GetResource(url, &resource)
	if err != nil {
		return CCServiceKey{}, err
	}

	serviceKey = resource.Entity
	serviceKey.ID = resource.Metadata.GUID
	return serviceKey, nil
}

// DeleteServiceKey -
func (sm *ServiceManager) DeleteServiceKey(serviceKeyID string) (err error) {
	err = sm.ccGateway.DeleteResource(sm.apiEndpoint, fmt.Sprintf("/v2/service_keys/%s", serviceKeyID))
	return err
}

// FindServiceKey -
func (sm *ServiceManager) FindServiceKey(name string, serviceInstanceID string) (serviceKey CCServiceKey, err error) {
	path := fmt.Sprintf("/v2/service_keys?q=%s", url.QueryEscape("name:"+name))

	var found bool

	apiErr := sm.ccGateway.ListPaginatedResources(
		sm.apiEndpoint,
		path,
		CCServiceKeyResource{},
		func(resource interface{}) bool {
			if sk, ok := resource.(CCServiceKeyResource); ok {
				if sk.Entity.ServiceGUID == serviceInstanceID {
					serviceKey = sk.Entity
					found = true
					return false
				}
			}
			return true
		})

	if apiErr != nil {
		err = apiErr
	} else {
		if !found {
			err = errors.NewModelNotFoundError("ServiceKey", name)
		}
	}

	return serviceKey, err
}

// FindSpaceService -
func (sm *ServiceManager) FindSpaceService(label string, spaceID string) (offering models.ServiceOffering, err error) {

	var offerings models.ServiceOfferings
	var count int

	offerings, err = sm.repo.FindServiceOfferingsForSpaceByLabel(spaceID, label)
	count = len(offerings)

	switch {
	case count < 1:
		err = fmt.Errorf("Service %s not found in space %s", label, spaceID)
	case count > 1:
		err = fmt.Errorf("Too many %s Services in space %s", label, spaceID)
	}

	offering = offerings[0]

	return offering, err
}

// FindServiceByName -
func (sm *ServiceManager) FindServiceByName(label string) (offering models.ServiceOffering, err error) {

	var offerings models.ServiceOfferings
	var count int

	offerings, err = sm.repo.FindServiceOfferingsByLabel(label)
	count = len(offerings)

	switch {
	case count < 1:
		err = fmt.Errorf("Service %s not found", label)
	case count > 1:
		err = fmt.Errorf("Too many %s Services", label)
	}

	if len(offerings) > 0 {
		offering = offerings[0]
	} else {
		err = fmt.Errorf("Service %s not found", label)
	}

	return offering, err
}

// GetServicePlans -
func (sm *ServiceManager) GetServicePlans(serviceID string) (servicePlans map[string]interface{}, err error) {

	servicePlans = make(map[string]interface{})

	err = sm.ccGateway.ListPaginatedResources(
		sm.apiEndpoint,
		fmt.Sprintf("/v2/services/%s/service_plans", serviceID),
		resources.ServicePlanResource{},
		func(resource interface{}) bool {
			if sp, ok := resource.(resources.ServicePlanResource); ok {
				servicePlans[strings.Replace(sp.Entity.Name, ".", "_", -1)] = sp.Metadata.GUID
			}
			return true
		})

	return servicePlans, err
}

// FindServicePlanID -
func (sm *ServiceManager) FindServicePlanID(serviceID string, plan string) (id string, err error) {
	servicePlans, err := sm.GetServicePlans(serviceID)
	if err != nil {
		return "", err
	}

	servicePlanID, ok := servicePlans[plan]
	if !ok {
		return "", fmt.Errorf("plan %s does not exist in service %s", plan, serviceID)
	}

	id = servicePlanID.(string)
	return id, nil
}

// ReadRouteServiceBindings -
func (sm *ServiceManager) ReadRouteServiceBindings(serviceInstanceID string) (routeIDs []string, err error) {
	path := fmt.Sprintf("/v2/service_instances/%s/routes", serviceInstanceID)
	err = sm.ccGateway.ListPaginatedResources(sm.apiEndpoint, path, CCServiceInstanceRouteResource{}, func(route interface{}) bool {
		r := route.(CCServiceInstanceRouteResource)
		routeIDs = append(routeIDs, r.Metadata.GUID)
		return true
	})
	if err != nil {
		return []string{}, err
	}

	return routeIDs, nil
}

// HasRouteServiceBinding -
func (sm *ServiceManager) HasRouteServiceBinding(serviceInstanceID, routeID string) (bool, error) {
	routes, err := sm.ReadRouteServiceBindings(serviceInstanceID)
	if err != nil {
		return false, err
	}
	for _, route := range routes {
		if route == routeID {
			return true, nil
		}
	}
	return false, nil
}

// CreateRouteServiceBinding -
func (sm *ServiceManager) CreateRouteServiceBinding(serviceID, routeID string, params interface{}) (err error) {
	path := fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceID, routeID)

	jsonBytes, err := json.Marshal(map[string]interface{}{
		"parameters": params,
	})
	if err != nil {
		return err
	}

	resource := CCServiceInstanceResource{}
	err = sm.ccGateway.UpdateResource(sm.apiEndpoint, path, bytes.NewReader(jsonBytes), &resource)
	return err
}

// DeleteRouteServiceBinding -
func (sm *ServiceManager) DeleteRouteServiceBinding(serviceID, routeID string) (err error) {
	path := fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceID, routeID)
	return sm.ccGateway.DeleteResource(sm.apiEndpoint, path)
}
