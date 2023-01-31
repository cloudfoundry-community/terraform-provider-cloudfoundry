package internal

import (
	"net/http"

	"github.com/tedsuo/rata"
)

// Naming convention:
//
// Method + non-parameter parts of the path
//
// If the request returns a single entity by GUID, use the singular (for example
// /v2/organizations/:organization_guid is GetOrganization).
//
// The const name should always be the const value + Request.
const (
	DeleteAppRequest                                     = "DeleteApp"
	DeleteBuildpackRequest                               = "DeleteBuildpackSpace"
	DeleteConfigRunningSecurityGroupRequest              = "DeleteConfigRunningSecurityGroup"
	DeleteConfigStagingSecurityGroupRequest              = "DeleteConfigStagingSecurityGroup"
	DeleteOrganizationRequest                            = "DeleteOrganization"
	DeleteOrganizationPrivateDomainRequest               = "DeleteOrganizationPrivateDomain"
	DeleteOrganizationQuotaDefinitionRequest             = "DeleteOrganizationQuotaDefinition"
	DeleteOrganizationManagerRequest                     = "DeleteOrganizationManager"
	DeleteOrganizationBillingManagerRequest              = "DeleteOrganizationBillingManager"
	DeleteOrganizationAuditorRequest                     = "DeleteOrganizationAuditorRequest"
	DeleteOrganizationUserRequest                        = "DeleteOrganizationUser"
	DeletePrivateDomainRequest                           = "DeletePrivateDomain"
	DeleteRouteAppRequest                                = "DeleteRouteApp"
	DeleteRouteRequest                                   = "DeleteRoute"
	DeleteRouteMappingRequest                            = "DeleteRouteMapping"
	DeleteSecurityGroupRequest                           = "DeleteSecurityGroup"
	DeleteSecurityGroupSpaceRequest                      = "DeleteSecurityGroupSpace"
	DeleteSecurityGroupStagingSpaceRequest               = "DeleteSecurityGroupStagingSpace"
	DeleteServiceBrokerRequest                           = "DeleteServiceBroker"
	DeleteServiceBindingRequest                          = "DeleteServiceBinding"
	DeleteServiceInstanceRequest                         = "DeleteServiceInstance"
	DeleteServicePlanRequest                             = "DeleteServicePlan"
	DeleteServicePlanVisibilityRequest                   = "DeleteServicePlanVisibility"
	DeleteServiceRequest                                 = "DeleteService"
	DeleteServiceBindingRouteRequest                     = "DeleteServiceBindingRoute"
	DeleteServiceKeyRequest                              = "DeleteServiceKey"
	DeleteSpaceRequest                                   = "DeleteSpace"
	DeleteSpaceAuditorRequest                            = "DeleteSpaceAuditor"
	DeleteSpaceDeveloperRequest                          = "DeleteSpaceDeveloper"
	DeleteSpaceManagerRequest                            = "DeleteSpaceManager"
	DeleteSpaceQuotaRequest                              = "DeleteSpaceQuota"
	DeleteSpaceQuotaDefinitionRequest                    = "DeleteSpaceQuotaDefinition"
	DeleteSharedDomainRequest                            = "DeleteSharedDomain"
	DeleteSpaceUnmappedRoutesRequest                     = "DeleteUnmappedRoutes"
	DeleteUserProvidedServiceInstanceRequest             = "DeleteUserProvidedServiceInstance"
	DeleteUserProvidedServiceInstanceRoutesRequest       = "DeleteUserProvidedServiceInstanceRoutes"
	GetAppInstancesRequest                               = "GetAppInstances"
	GetAppRequest                                        = "GetApp"
	GetAppRoutesRequest                                  = "GetAppRoutes"
	GetAppsRequest                                       = "GetApps"
	GetAppStatsRequest                                   = "GetAppStats"
	GetBuildpacksRequest                                 = "GetBuildpacks"
	GetBuildpackRequest                                  = "GetBuildpack"
	GetConfigFeatureFlagsRequest                         = "GetConfigFeatureFlags"
	GetConfigEnvVarGroupRunningRequest                   = "GetConfigEnvVarGroupRunning"
	GetConfigEnvVarGroupStagingRequest                   = "GetConfigEnvVarGroupStaging"
	GetConfigRunningSecurityGroupsRequest                = "GetConfigRunningSecurityGroups"
	GetConfigStagingSecurityGroupsRequest                = "GetConfigStagingSecurityGroups"
	GetEventsRequest                                     = "GetEvents"
	GetInfoRequest                                       = "GetInfo"
	GetJobRequest                                        = "GetJob"
	GetOrganizationPrivateDomainsRequest                 = "GetOrganizationPrivateDomains"
	GetOrganizationQuotaDefinitionsRequest               = "GetOrganizationQuotaDefinitions"
	GetOrganizationQuotaDefinitionRequest                = "GetOrganizationQuotaDefinition"
	GetOrganizationAuditorsRequest                       = "GetOrganizationAuditors"
	GetOrganizationManagersRequest                       = "GetOrganizationManagers"
	GetOrganizationBillingManagersRequest                = "GetOrganizationBillingManagers"
	GetOrganizationUsersRequest                          = "GetOrganizationUsers"
	GetOrganizationRequest                               = "GetOrganization"
	GetOrganizationsRequest                              = "GetOrganizations"
	GetPrivateDomainRequest                              = "GetPrivateDomain"
	GetPrivateDomainsRequest                             = "GetPrivateDomains"
	GetRouteAppsRequest                                  = "GetRouteApps"
	GetRouteMappingRequest                               = "GetRouteMapping"
	GetRouteMappingsRequest                              = "GetRouteMappings"
	GetRouteRequest                                      = "GetRoute"
	GetRouteReservedDeprecatedRequest                    = "GetRouteReservedDeprecated"
	GetRouteReservedRequest                              = "GetRouteReserved"
	GetRouteRouteMappingsRequest                         = "GetRouteRouteMappings"
	GetRoutesRequest                                     = "GetRoutes"
	GetSecurityGroupSpacesRequest                        = "GetSecurityGroupSpaces"
	GetSecurityGroupsRequest                             = "GetSecurityGroups"
	GetSecurityGroupStagingSpacesRequest                 = "GetSecurityGroupStagingSpaces"
	GetServiceBindingRequest                             = "GetServiceBinding"
	GetServiceBindingsRequest                            = "GetServiceBindings"
	GetServiceBindingRouteRequest                        = "GetServiceBindingRoute"
	GetServiceBindingRoutesRequest                       = "GetServiceBindingRoutes"
	GetServiceBrokersRequest                             = "GetServiceBrokers"
	GetServiceInstanceRequest                            = "GetServiceInstance"
	GetServiceInstanceServiceBindingsRequest             = "GetServiceInstanceServiceBindings"
	GetServiceInstanceSharedFromRequest                  = "GetServiceInstanceSharedFrom"
	GetServiceInstanceSharedToRequest                    = "GetServiceInstanceSharedTo"
	GetServiceInstancesRequest                           = "GetServiceInstances"
	GetServicePlanRequest                                = "GetServicePlan"
	GetServicePlansRequest                               = "GetServicePlans"
	GetServicePlanVisibilitiesRequest                    = "GetServicePlanVisibilities"
	GetServiceRequest                                    = "GetService"
	GetServicesRequest                                   = "GetServices"
	GetServiceKeyRequest                                 = "GetServiceKey"
	GetSharedDomainRequest                               = "GetSharedDomain"
	GetSharedDomainsRequest                              = "GetSharedDomains"
	GetSpaceRequest                                      = "GetSpace"
	GetSpaceQuotaDefinitionsRequest                      = "GetSpaceQuotaDefinitions"
	GetOrganizationSpaceQuotasRequest                    = "GetOrganizationSpaceQuotas"
	GetSecurityGroupRequest                              = "GetSecurityGroup"
	GetServiceBrokerRequest                              = "GetServiceBroker"
	GetServicePlanVisibilityRequest                      = "GetServicePlanVisibility"
	GetSpaceAuditorsRequest                              = "GetSpaceAuditors"
	GetSpaceManagersRequest                              = "GetSpaceManagers"
	GetSpaceDevelopersRequest                            = "GetSpaceDevelopers"
	GetSpaceQuotaDefinitionRequest                       = "GetSpaceQuotaDefinition"
	GetSpaceRoutesRequest                                = "GetSpaceRoutes"
	GetSpaceSecurityGroupsRequest                        = "GetSpaceSecurityGroups"
	GetSpaceServicesRequest                              = "GetSpaceServices"
	GetSpaceServiceInstancesRequest                      = "GetSpaceServiceInstances"
	GetSpaceSummaryRequest                               = "GetSpaceSummary"
	GetSpacesRequest                                     = "GetSpaces"
	GetSpaceStagingSecurityGroupsRequest                 = "GetSpaceStagingSecurityGroups"
	PostSpaceQuotaDefinitionsRequest                     = "PostSpaceQuotaDefinitions"
	GetStackRequest                                      = "GetStack"
	GetStacksRequest                                     = "GetStacks"
	GetUserProvidedServiceInstanceServiceBindingsRequest = "GetUserProvidedServiceInstanceServiceBindings"
	GetUserProvidedServiceInstanceRequest                = "GetUserProvidedServiceInstance"
	GetUserProvidedServiceInstancesRequest               = "GetUserProvidedServiceInstances"
	GetUsersRequest                                      = "GetUsers"
	GetUserOrganizationsRequest                          = "GetUserOrganizations"
	GetUserProvidedServiceInstanceRoutesRequest          = "GetUserProvidedServiceInstanceRoutes"
	GetUserSpacesRequest                                 = "GetUserSpaces"
	PostAppRequest                                       = "PostApp"
	PostAppRestageRequest                                = "PostAppRestage"
	PostBuildpackRequest                                 = "PostBuildpack"
	PutConfigFeatureFlagsRequest                         = "PutConfigFeatureFlags"
	PostOrganizationRequest                              = "PostOrganization"
	PostOrganizationQuotaDefinitionsRequest              = "PostOrganizationQuotaDefinitions"
	PostPrivateDomainRequest                             = "PostPrivateDomain"
	PostRouteRequest                                     = "PostRoute"
	PostRouteMappingsRequest                             = "PostRouteMappings"
	PostSecurityGroupsRequest                            = "PostSecurityGroups"
	PostSecurityGroupRequest                             = "PostSecurityGroup"
	PostServiceBindingRequest                            = "PostServiceBinding"
	PostServiceInstancesRequest                          = "PostServiceInstance"
	PostSharedDomainRequest                              = "PostSharedDomain"
	PostServicesRequest                                  = "PostServicesRequest"
	PostServiceBrokerRequest                             = "PostServiceBroker"
	PostServiceKeyRequest                                = "PostServiceKey"
	GetServiceKeysRequest                                = "GetServiceKeys"
	PostServicePlanVisibilityRequest                     = "PostServicePlanVisibility"
	PostSpaceRequest                                     = "PostSpace"
	PostUserRequest                                      = "PostUser"
	PostUserProvidedServiceInstancesRequest              = "PostUserProvidedServiceInstances"
	PutAppBitsRequest                                    = "PutAppBits"
	PutAppRequest                                        = "PutApp"
	PutBuildpackRequest                                  = "PutBuildpack"
	PutBuildpackBitsRequest                              = "PutBuildpackBits"
	PutConfigRunningSecurityGroupRequest                 = "PutConfigRunningSecurityGroup"
	PutConfigEnvVarGroupRunningRequest                   = "PutConfigEnvVarGroupRunning"
	PutConfigEnvVarGroupStagingRequest                   = "PutConfigEnvVarGroupStaging"
	PutConfigStagingSecurityGroupRequest                 = "PutConfigStagingSecurityGroup"
	PutDropletRequest                                    = "PutDroplet"
	PutOrganizationBillingManagerByUsernameRequest       = "PutOrganizationBillingManagerByUsername"
	DeleteOrganizationBillingManagerByUsernameRequest    = "DeleteOrganizationBillingManagerByUsername"
	PutOrganizationAuditorByUsernameRequest              = "PutOrganizationAuditorByUsername"
	DeleteOrganizationAuditorByUsernameRequest           = "DeleteOrganizationAuditorByUsername"
	PutOrganizationManagerByUsernameRequest              = "PutOrganizationManagerByUsername"
	DeleteOrganizationManagerByUsernameRequest           = "DeleteOrganizationManagerByUsername"
	PutOrganizationManagerRequest                        = "PutOrganizationManager"
	PutOrganizationUserRequest                           = "PutOrganizationUser"
	PutOrganizationAuditorRequest                        = "PutOrganizationAuditorRequest"
	PutOrganizationBillingManagerRequest                 = "PutOrganizationBillingManager"
	PutOrganizationRequest                               = "PutOrganizationRequest"
	PutOrganizationUserByUsernameRequest                 = "PutOrganizationUserByUsername"
	DeleteOrganizationUserByUsernameRequest              = "DeleteOrganizationUserByUsername"
	PutOrganizationPrivateDomainRequest                  = "PutOrganizationPrivateDomain"
	PutOrganizationQuotaDefinitionRequest                = "PutOrganizationQuotaDefinition"
	PutResourceMatchRequest                              = "PutResourceMatch"
	PutRouteAppRequest                                   = "PutRouteApp"
	PutRouteRequest                                      = "PutRoute"
	PutServiceRequest                                    = "PutServiceRequest"
	PutServiceBrokerRequest                              = "PutServiceBroker"
	PutServiceBindingRoutesRequest                       = "PutServiceBindingRoutes"
	PutServiceInstanceRequest                            = "PutServiceInstance"
	PutServicePlanRequest                                = "PutServicePlan"
	PutServicePlanVisibilityRequest                      = "PutServicePlanVisibility"
	PutSpaceRequest                                      = "PutSpace"
	PutSpaceQuotaRequest                                 = "PutSpaceQuotaRequest"
	PutSpaceAuditorByUsernameRequest                     = "PutSpaceAuditorByUsername"
	DeleteSpaceAuditorByUsernameRequest                  = "DeleteSpaceAuditorByUsername"
	PutSpaceDeveloperRequest                             = "PutSpaceDeveloper"
	PutSpaceDeveloperByUsernameRequest                   = "PutSpaceDeveloperByUsername"
	DeleteSpaceDeveloperByUsernameRequest                = "DeleteSpaceDeveloperByUsername"
	PutSpaceManagerRequest                               = "PutSpaceManager"
	PutSpaceAuditorRequest                               = "PutSpaceAuditor"
	PutSpaceManagerByUsernameRequest                     = "PutSpaceManagerByUsername"
	DeleteSpaceManagerByUsernameRequest                  = "DeleteSpaceManagerByUsername"
	PutSpaceQuotaDefinitionRequest                       = "PutSpaceQuotaDefinition"
	PutSecurityGroupSpaceRequest                         = "PutSecurityGroupSpace"
	PutSecurityGroupStagingSpaceRequest                  = "PutSecurityGroupStagingSpace"
	PutUserProvidedServiceInstanceRequest                = "PutUserProvidedServiceInstance"
)

// APIRoutes is a list of routes used by the rata library to construct request
// URLs.
var APIRoutes = rata.Routes{
	{Path: "/v2/apps", Method: http.MethodGet, Name: GetAppsRequest},
	{Path: "/v2/apps", Method: http.MethodPost, Name: PostAppRequest},
	{Path: "/v2/apps/:app_guid", Method: http.MethodGet, Name: GetAppRequest},
	{Path: "/v2/apps/:app_guid", Method: http.MethodDelete, Name: DeleteAppRequest},
	{Path: "/v2/apps/:app_guid", Method: http.MethodPut, Name: PutAppRequest},
	{Path: "/v2/apps/:app_guid/bits", Method: http.MethodPut, Name: PutAppBitsRequest},
	{Path: "/v2/apps/:app_guid/droplet/upload", Method: http.MethodPut, Name: PutDropletRequest},
	{Path: "/v2/apps/:app_guid/instances", Method: http.MethodGet, Name: GetAppInstancesRequest},
	{Path: "/v2/apps/:app_guid/restage", Method: http.MethodPost, Name: PostAppRestageRequest},
	{Path: "/v2/apps/:app_guid/routes", Method: http.MethodGet, Name: GetAppRoutesRequest},
	{Path: "/v2/apps/:app_guid/stats", Method: http.MethodGet, Name: GetAppStatsRequest},
	{Path: "/v2/buildpacks", Method: http.MethodPost, Name: PostBuildpackRequest},
	{Path: "/v2/buildpacks", Method: http.MethodGet, Name: GetBuildpacksRequest},
	{Path: "/v2/buildpacks/:buildpack_guid", Method: http.MethodGet, Name: GetBuildpackRequest},
	{Path: "/v2/buildpacks/:buildpack_guid", Method: http.MethodDelete, Name: DeleteBuildpackRequest},
	{Path: "/v2/buildpacks/:buildpack_guid", Method: http.MethodPut, Name: PutBuildpackRequest},
	{Path: "/v2/buildpacks/:buildpack_guid/bits", Method: http.MethodPut, Name: PutBuildpackBitsRequest},
	{Path: "/v2/config/feature_flags", Method: http.MethodGet, Name: GetConfigFeatureFlagsRequest},
	{Path: "/v2/config/feature_flags/:name", Method: http.MethodPut, Name: PutConfigFeatureFlagsRequest},
	{Path: "/v2/config/running_security_groups", Method: http.MethodGet, Name: GetConfigRunningSecurityGroupsRequest},
	{Path: "/v2/config/running_security_groups/:security_group_guid", Method: http.MethodPut, Name: PutConfigRunningSecurityGroupRequest},
	{Path: "/v2/config/running_security_groups/:security_group_guid", Method: http.MethodDelete, Name: DeleteConfigRunningSecurityGroupRequest},
	{Path: "/v2/config/staging_security_groups", Method: http.MethodGet, Name: GetConfigStagingSecurityGroupsRequest},
	{Path: "/v2/config/staging_security_groups/:security_group_guid", Method: http.MethodPut, Name: PutConfigStagingSecurityGroupRequest},
	{Path: "/v2/config/staging_security_groups/:security_group_guid", Method: http.MethodDelete, Name: DeleteConfigStagingSecurityGroupRequest},
	{Path: "/v2/config/environment_variable_groups/running", Method: http.MethodGet, Name: GetConfigEnvVarGroupRunningRequest},
	{Path: "/v2/config/environment_variable_groups/running", Method: http.MethodPut, Name: PutConfigEnvVarGroupRunningRequest},
	{Path: "/v2/config/environment_variable_groups/staging", Method: http.MethodGet, Name: GetConfigEnvVarGroupStagingRequest},
	{Path: "/v2/config/environment_variable_groups/staging", Method: http.MethodPut, Name: PutConfigEnvVarGroupStagingRequest},
	{Path: "/v2/events", Method: http.MethodGet, Name: GetEventsRequest},
	{Path: "/v2/info", Method: http.MethodGet, Name: GetInfoRequest},
	{Path: "/v2/jobs/:job_guid", Method: http.MethodGet, Name: GetJobRequest},
	{Path: "/v2/organizations", Method: http.MethodGet, Name: GetOrganizationsRequest},
	{Path: "/v2/organizations", Method: http.MethodPost, Name: PostOrganizationRequest},
	{Path: "/v2/organizations/:organization_guid", Method: http.MethodDelete, Name: DeleteOrganizationRequest},
	{Path: "/v2/organizations/:organization_guid", Method: http.MethodGet, Name: GetOrganizationRequest},
	{Path: "/v2/organizations/:organization_guid", Method: http.MethodPut, Name: PutOrganizationRequest},
	{Path: "/v2/organizations/:organization_guid/billing_managers", Method: http.MethodPut, Name: PutOrganizationBillingManagerByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/billing_managers/remove", Method: http.MethodPost, Name: DeleteOrganizationBillingManagerByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/auditors", Method: http.MethodPut, Name: PutOrganizationAuditorByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/auditors/remove", Method: http.MethodPost, Name: DeleteOrganizationAuditorByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/managers", Method: http.MethodPut, Name: PutOrganizationManagerByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/managers/remove", Method: http.MethodPost, Name: DeleteOrganizationManagerByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/managers", Method: http.MethodGet, Name: GetOrganizationManagersRequest},
	{Path: "/v2/organizations/:organization_guid/users", Method: http.MethodGet, Name: GetOrganizationUsersRequest},
	{Path: "/v2/organizations/:organization_guid/auditors", Method: http.MethodGet, Name: GetOrganizationAuditorsRequest},
	{Path: "/v2/organizations/:organization_guid/billing_managers", Method: http.MethodGet, Name: GetOrganizationBillingManagersRequest},
	{Path: "/v2/organizations/:organization_guid/managers/:manager_guid", Method: http.MethodPut, Name: PutOrganizationManagerRequest},
	{Path: "/v2/organizations/:organization_guid/billing_managers/:billing_manager_guid", Method: http.MethodPut, Name: PutOrganizationBillingManagerRequest},
	{Path: "/v2/organizations/:organization_guid/auditors/:auditor_guid", Method: http.MethodPut, Name: PutOrganizationAuditorRequest},
	{Path: "/v2/organizations/:organization_guid/managers/:manager_guid", Method: http.MethodDelete, Name: DeleteOrganizationManagerRequest},
	{Path: "/v2/organizations/:organization_guid/billing_managers/:billing_manager_guid", Method: http.MethodDelete, Name: DeleteOrganizationBillingManagerRequest},
	{Path: "/v2/organizations/:organization_guid/auditors/:auditor_guid", Method: http.MethodDelete, Name: DeleteOrganizationAuditorRequest},
	{Path: "/v2/organizations/:organization_guid/private_domains", Method: http.MethodGet, Name: GetOrganizationPrivateDomainsRequest},
	{Path: "/v2/organizations/:organization_guid/private_domains/:private_domain_guid", Method: http.MethodPut, Name: PutOrganizationPrivateDomainRequest},
	{Path: "/v2/organizations/:organization_guid/private_domains/:private_domain_guid", Method: http.MethodDelete, Name: DeleteOrganizationPrivateDomainRequest},
	{Path: "/v2/organizations/:organization_guid/users", Method: http.MethodPut, Name: PutOrganizationUserByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/users/remove", Method: http.MethodPost, Name: DeleteOrganizationUserByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/users/:user_guid", Method: http.MethodPut, Name: PutOrganizationUserRequest},
	{Path: "/v2/organizations/:organization_guid/users/:user_guid", Method: http.MethodDelete, Name: DeleteOrganizationUserRequest},
	{Path: "/v2/private_domains", Method: http.MethodGet, Name: GetPrivateDomainsRequest},
	{Path: "/v2/private_domains", Method: http.MethodPost, Name: PostPrivateDomainRequest},
	{Path: "/v2/private_domains/:private_domain_guid", Method: http.MethodGet, Name: GetPrivateDomainRequest},
	{Path: "/v2/private_domains/:private_domain_guid", Method: http.MethodDelete, Name: DeletePrivateDomainRequest},
	{Path: "/v2/quota_definitions/:organization_quota_guid", Method: http.MethodGet, Name: GetOrganizationQuotaDefinitionRequest},
	{Path: "/v2/quota_definitions/:organization_quota_guid", Method: http.MethodPut, Name: PutOrganizationQuotaDefinitionRequest},
	{Path: "/v2/quota_definitions/:organization_quota_guid", Method: http.MethodDelete, Name: DeleteOrganizationQuotaDefinitionRequest},
	{Path: "/v2/quota_definitions", Method: http.MethodGet, Name: GetOrganizationQuotaDefinitionsRequest},
	{Path: "/v2/quota_definitions", Method: http.MethodPost, Name: PostOrganizationQuotaDefinitionsRequest},
	{Path: "/v2/resource_match", Method: http.MethodPut, Name: PutResourceMatchRequest},
	{Path: "/v2/route_mappings", Method: http.MethodGet, Name: GetRouteMappingsRequest},
	{Path: "/v2/route_mappings", Method: http.MethodPost, Name: PostRouteMappingsRequest},
	{Path: "/v2/route_mappings/:route_mapping_guid", Method: http.MethodGet, Name: GetRouteMappingRequest},
	{Path: "/v2/route_mappings/:route_mapping_guid", Method: http.MethodDelete, Name: DeleteRouteMappingRequest},
	{Path: "/v2/routes", Method: http.MethodGet, Name: GetRoutesRequest},
	{Path: "/v2/routes", Method: http.MethodPost, Name: PostRouteRequest},
	{Path: "/v2/routes/:route_guid", Method: http.MethodDelete, Name: DeleteRouteRequest},
	{Path: "/v2/routes/:route_guid", Method: http.MethodGet, Name: GetRouteRequest},
	{Path: "/v2/routes/:route_guid", Method: http.MethodPut, Name: PutRouteRequest},
	{Path: "/v2/routes/:route_guid/apps", Method: http.MethodGet, Name: GetRouteAppsRequest},
	{Path: "/v2/routes/:route_guid/apps/:app_guid", Method: http.MethodDelete, Name: DeleteRouteAppRequest},
	{Path: "/v2/routes/:route_guid/apps/:app_guid", Method: http.MethodPut, Name: PutRouteAppRequest},
	{Path: "/v2/routes/:route_guid/route_mappings", Method: http.MethodGet, Name: GetRouteRouteMappingsRequest},
	{Path: "/v2/routes/reserved/domain/:domain_guid", Method: http.MethodGet, Name: GetRouteReservedRequest},
	{Path: "/v2/routes/reserved/domain/:domain_guid/host/:host", Method: http.MethodGet, Name: GetRouteReservedDeprecatedRequest},
	{Path: "/v2/security_groups", Method: http.MethodGet, Name: GetSecurityGroupsRequest},
	{Path: "/v2/security_groups", Method: http.MethodPost, Name: PostSecurityGroupsRequest},
	{Path: "/v2/security_groups/:security_group_guid", Method: http.MethodPut, Name: PostSecurityGroupRequest},
	{Path: "/v2/security_groups/:security_group_guid", Method: http.MethodGet, Name: GetSecurityGroupRequest},
	{Path: "/v2/security_groups/:security_group_guid", Method: http.MethodDelete, Name: DeleteSecurityGroupRequest},
	{Path: "/v2/security_groups/:security_group_guid/spaces", Method: http.MethodGet, Name: GetSecurityGroupSpacesRequest},
	{Path: "/v2/security_groups/:security_group_guid/spaces/:space_guid", Method: http.MethodDelete, Name: DeleteSecurityGroupSpaceRequest},
	{Path: "/v2/security_groups/:security_group_guid/spaces/:space_guid", Method: http.MethodPut, Name: PutSecurityGroupSpaceRequest},
	{Path: "/v2/security_groups/:security_group_guid/staging_spaces", Method: http.MethodGet, Name: GetSecurityGroupStagingSpacesRequest},
	{Path: "/v2/security_groups/:security_group_guid/staging_spaces/:space_guid", Method: http.MethodDelete, Name: DeleteSecurityGroupStagingSpaceRequest},
	{Path: "/v2/security_groups/:security_group_guid/staging_spaces/:space_guid", Method: http.MethodPut, Name: PutSecurityGroupStagingSpaceRequest},
	{Path: "/v2/service_bindings", Method: http.MethodGet, Name: GetServiceBindingsRequest},
	{Path: "/v2/service_bindings", Method: http.MethodPost, Name: PostServiceBindingRequest},
	{Path: "/v2/service_bindings/:service_binding_guid", Method: http.MethodDelete, Name: DeleteServiceBindingRequest},
	{Path: "/v2/service_bindings/:service_binding_guid", Method: http.MethodGet, Name: GetServiceBindingRequest},
	{Path: "/v2/service_instances/:service_guid/routes", Method: http.MethodGet, Name: GetServiceBindingRoutesRequest},
	{Path: "/v2/service_instances/:service_guid/routes/:route_guid", Method: http.MethodPut, Name: PutServiceBindingRoutesRequest},
	{Path: "/v2/service_instances/:service_guid/routes/:route_guid", Method: http.MethodGet, Name: GetServiceBindingRouteRequest},
	{Path: "/v2/service_instances/:service_guid/routes/:route_guid", Method: http.MethodDelete, Name: DeleteServiceBindingRouteRequest},
	{Path: "/v2/service_brokers", Method: http.MethodGet, Name: GetServiceBrokersRequest},
	{Path: "/v2/service_brokers", Method: http.MethodPost, Name: PostServiceBrokerRequest},
	{Path: "/v2/service_brokers/:service_broker_guid", Method: http.MethodPut, Name: PutServiceBrokerRequest},
	{Path: "/v2/service_brokers/:service_broker_guid", Method: http.MethodDelete, Name: DeleteServiceBrokerRequest},
	{Path: "/v2/service_brokers/:service_broker_guid", Method: http.MethodGet, Name: GetServiceBrokerRequest},
	{Path: "/v2/service_instances", Method: http.MethodGet, Name: GetServiceInstancesRequest},
	{Path: "/v2/service_instances", Method: http.MethodPost, Name: PostServiceInstancesRequest},
	{Path: "/v2/service_instances/:service_instance_guid", Method: http.MethodGet, Name: GetServiceInstanceRequest},
	{Path: "/v2/service_instances/:service_instance_guid", Method: http.MethodPut, Name: PutServiceInstanceRequest},
	{Path: "/v2/service_instances/:service_instance_guid", Method: http.MethodDelete, Name: DeleteServiceInstanceRequest},
	{Path: "/v2/service_instances/:service_instance_guid/service_bindings", Method: http.MethodGet, Name: GetServiceInstanceServiceBindingsRequest},
	{Path: "/v2/service_instances/:service_instance_guid/shared_from", Method: http.MethodGet, Name: GetServiceInstanceSharedFromRequest},
	{Path: "/v2/service_instances/:service_instance_guid/shared_to", Method: http.MethodGet, Name: GetServiceInstanceSharedToRequest},
	{Path: "/v2/service_keys", Method: http.MethodPost, Name: PostServiceKeyRequest},
	{Path: "/v2/service_keys", Method: http.MethodGet, Name: GetServiceKeysRequest},
	{Path: "/v2/service_keys/:service_key_guid", Method: http.MethodGet, Name: GetServiceKeyRequest},
	{Path: "/v2/service_keys/:service_key_guid", Method: http.MethodDelete, Name: DeleteServiceKeyRequest},
	{Path: "/v2/service_plan_visibilities", Method: http.MethodGet, Name: GetServicePlanVisibilitiesRequest},
	{Path: "/v2/service_plan_visibilities", Method: http.MethodPost, Name: PostServicePlanVisibilityRequest},
	{Path: "/v2/service_plan_visibilities/:service_plan_visibility_guid", Method: http.MethodDelete, Name: DeleteServicePlanVisibilityRequest},
	{Path: "/v2/service_plan_visibilities/:service_plan_visibility_guid", Method: http.MethodGet, Name: GetServicePlanVisibilityRequest},
	{Path: "/v2/service_plan_visibilities/:service_plan_visibility_guid", Method: http.MethodPut, Name: PutServicePlanVisibilityRequest},
	{Path: "/v2/service_plans", Method: http.MethodGet, Name: GetServicePlansRequest},
	{Path: "/v2/service_plans/:service_plan_guid", Method: http.MethodPut, Name: PutServicePlanRequest},
	{Path: "/v2/service_plans/:service_plan_guid", Method: http.MethodDelete, Name: DeleteServicePlanRequest},
	{Path: "/v2/service_plans/:service_plan_guid", Method: http.MethodGet, Name: GetServicePlanRequest},
	{Path: "/v2/services", Method: http.MethodGet, Name: GetServicesRequest},
	{Path: "/v2/services", Method: http.MethodPost, Name: PostServicesRequest},
	{Path: "/v2/services/:service_guid", Method: http.MethodGet, Name: GetServiceRequest},
	{Path: "/v2/services/:service_guid", Method: http.MethodPut, Name: PutServiceRequest},
	{Path: "/v2/services/:service_guid", Method: http.MethodDelete, Name: DeleteServiceRequest},
	{Path: "/v2/shared_domains", Method: http.MethodGet, Name: GetSharedDomainsRequest},
	{Path: "/v2/shared_domains", Method: http.MethodPost, Name: PostSharedDomainRequest},
	{Path: "/v2/shared_domains/:shared_domain_guid", Method: http.MethodGet, Name: GetSharedDomainRequest},
	{Path: "/v2/shared_domains/:shared_domain_guid", Method: http.MethodDelete, Name: DeleteSharedDomainRequest},
	{Path: "/v2/organizations/:space_guid/space_quota_definitions", Method: http.MethodGet, Name: GetOrganizationSpaceQuotasRequest},
	{Path: "/v2/space_quota_definitions/:space_quota_guid/spaces/:space_guid", Method: http.MethodPut, Name: PutSpaceQuotaRequest},
	{Path: "/v2/space_quota_definitions/:space_quota_guid/spaces/:space_guid", Method: http.MethodDelete, Name: DeleteSpaceQuotaRequest},
	{Path: "/v2/space_quota_definitions", Method: http.MethodGet, Name: GetSpaceQuotaDefinitionsRequest},
	{Path: "/v2/space_quota_definitions", Method: http.MethodPost, Name: PostSpaceQuotaDefinitionsRequest},
	{Path: "/v2/space_quota_definitions/:space_quota_guid", Method: http.MethodGet, Name: GetSpaceQuotaDefinitionRequest},
	{Path: "/v2/space_quota_definitions/:space_quota_guid", Method: http.MethodDelete, Name: DeleteSpaceQuotaDefinitionRequest},
	{Path: "/v2/space_quota_definitions/:space_quota_guid", Method: http.MethodPut, Name: PutSpaceQuotaDefinitionRequest},
	{Path: "/v2/spaces/:space_guid/summary", Method: http.MethodGet, Name: GetSpaceSummaryRequest},
	{Path: "/v2/spaces", Method: http.MethodGet, Name: GetSpacesRequest},
	{Path: "/v2/spaces", Method: http.MethodPost, Name: PostSpaceRequest},
	{Path: "/v2/spaces/:space_guid/developers", Method: http.MethodPut, Name: PutSpaceDeveloperByUsernameRequest},
	{Path: "/v2/spaces/:space_guid/developers/remove", Method: http.MethodPost, Name: DeleteSpaceDeveloperByUsernameRequest},
	{Path: "/v2/spaces/:space_guid/developers", Method: http.MethodGet, Name: GetSpaceDevelopersRequest},
	{Path: "/v2/spaces/:space_guid/auditors", Method: http.MethodPut, Name: PutSpaceAuditorByUsernameRequest},
	{Path: "/v2/spaces/:space_guid/auditors/remove", Method: http.MethodPost, Name: DeleteSpaceAuditorByUsernameRequest},
	{Path: "/v2/spaces/:space_guid/auditors", Method: http.MethodGet, Name: GetSpaceAuditorsRequest},
	{Path: "/v2/spaces/:space_guid/managers", Method: http.MethodGet, Name: GetSpaceManagersRequest},
	{Path: "/v2/spaces/:space_guid/developers/:developer_guid", Method: http.MethodPut, Name: PutSpaceDeveloperRequest},
	{Path: "/v2/spaces/:space_guid/auditors/:auditor_guid", Method: http.MethodPut, Name: PutSpaceAuditorRequest},
	{Path: "/v2/spaces/:space_guid/developers/:developer_guid", Method: http.MethodDelete, Name: DeleteSpaceDeveloperRequest},
	{Path: "/v2/spaces/:space_guid/auditors/:auditor_guid", Method: http.MethodDelete, Name: DeleteSpaceAuditorRequest},
	{Path: "/v2/spaces/:guid/service_instances", Method: http.MethodGet, Name: GetSpaceServiceInstancesRequest},
	{Path: "/v2/spaces/:space_guid/services", Method: http.MethodGet, Name: GetSpaceServicesRequest},
	{Path: "/v2/spaces/:space_guid", Method: http.MethodDelete, Name: DeleteSpaceRequest},
	{Path: "/v2/spaces/:space_guid", Method: http.MethodGet, Name: GetSpaceRequest},
	{Path: "/v2/spaces/:space_guid", Method: http.MethodPut, Name: PutSpaceRequest},
	{Path: "/v2/spaces/:space_guid/routes", Method: http.MethodGet, Name: GetSpaceRoutesRequest},
	{Path: "/v2/spaces/:space_guid/security_groups", Method: http.MethodGet, Name: GetSpaceSecurityGroupsRequest},
	{Path: "/v2/spaces/:space_guid/staging_security_groups", Method: http.MethodGet, Name: GetSpaceStagingSecurityGroupsRequest},
	{Path: "/v2/spaces/:space_guid/managers", Method: http.MethodPut, Name: PutSpaceManagerByUsernameRequest},
	{Path: "/v2/spaces/:space_guid/managers/remove", Method: http.MethodPost, Name: DeleteSpaceManagerByUsernameRequest},
	{Path: "/v2/spaces/:space_guid/managers/:manager_guid", Method: http.MethodPut, Name: PutSpaceManagerRequest},
	{Path: "/v2/spaces/:space_guid/managers/:manager_guid", Method: http.MethodDelete, Name: DeleteSpaceManagerRequest},
	{Path: "/v2/spaces/:space_guid/unmapped_routes", Method: http.MethodDelete, Name: DeleteSpaceUnmappedRoutesRequest},
	{Path: "/v2/stacks", Method: http.MethodGet, Name: GetStacksRequest},
	{Path: "/v2/stacks/:stack_guid", Method: http.MethodGet, Name: GetStackRequest},
	{Path: "/v2/user_provided_service_instances", Method: http.MethodGet, Name: GetUserProvidedServiceInstancesRequest},
	{Path: "/v2/user_provided_service_instances", Method: http.MethodPost, Name: PostUserProvidedServiceInstancesRequest},
	{Path: "/v2/user_provided_service_instances/:user_provided_service_instance_guid", Method: http.MethodPut, Name: PutUserProvidedServiceInstanceRequest},
	{Path: "/v2/user_provided_service_instances/:user_provided_service_instance_guid", Method: http.MethodGet, Name: GetUserProvidedServiceInstanceRequest},
	{Path: "/v2/user_provided_service_instances/:user_provided_service_instance_guid", Method: http.MethodDelete, Name: DeleteUserProvidedServiceInstanceRequest},
	{Path: "/v2/user_provided_service_instances/:user_provided_service_instance_guid/service_bindings", Method: http.MethodGet, Name: GetUserProvidedServiceInstanceServiceBindingsRequest},
	{Path: "/v2/user_provided_service_instances/:service_guid/routes", Method: http.MethodGet, Name: GetUserProvidedServiceInstanceRoutesRequest},
	{Path: "/v2/user_provided_service_instances/:service_guid/routes/:route_guid", Method: http.MethodDelete, Name: DeleteUserProvidedServiceInstanceRoutesRequest},
	{Path: "/v2/users", Method: http.MethodPost, Name: PostUserRequest},
	{Path: "/v2/users", Method: http.MethodGet, Name: GetUsersRequest},
	{Path: "/v2/users/:user_guid/organizations", Method: http.MethodGet, Name: GetUserOrganizationsRequest},
	{Path: "/v2/users/:user_guid/spaces", Method: http.MethodGet, Name: GetUserSpacesRequest},
}
