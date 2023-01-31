package internal

// When adding a resource, also add it to the api/cloudcontroller/ccv3/ccv3_suite_test.go resources response
const (
	AppsResource              = "apps"
	BuildpacksResource        = "buildpacks"
	BuildsResource            = "builds"
	DeploymentsResource       = "deployments"
	DomainsResource           = "domains"
	DropletsResource          = "droplets"
	FeatureFlagsResource      = "feature_flags"
	IsolationSegmentsResource = "isolation_segments"
	OrgsResource              = "organizations"
	PackagesResource          = "packages"
	ProcessesResource         = "processes"
	ResourceMatches           = "resource_matches"
	ServiceInstancesResource  = "service_instances"
	SpacesResource            = "spaces"
	StacksResource            = "stacks"
	TasksResource             = "tasks"

	ServiceOfferingsResource = "service_offerings"

	ServicePlansResource = "service_plans"

	RoutesResource = "routes"

	//v3 service credential binding
	ServiceCredentialBindingsResource = "service_credential_bindings"

	//v3 environment variable
	EnvironmentVariableGroupsResource = "environment_variable_groups"

	// v3 organization quota
	OrgQuotasResource = "organization_quotas"
)
