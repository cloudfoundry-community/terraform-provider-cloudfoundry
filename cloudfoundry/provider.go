package cloudfoundry

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

// Provider -
func Provider() *schema.Provider {

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_API_URL", ""),
			},
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_USER", "admin"),
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_PASSWORD", ""),
			},
			"sso_passcode": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_SSO_PASSCODE", ""),
			},
			"cf_client_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_CLIENT_ID", ""),
			},
			"cf_client_secret": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_CLIENT_SECRET", ""),
			},
			"uaa_client_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_UAA_CLIENT_ID", ""),
			},
			"uaa_client_secret": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_UAA_CLIENT_SECRET", "admin"),
			},
			"skip_ssl_validation": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_SKIP_SSL_VALIDATION", false),
			},
			"default_quota_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the default quota",
				DefaultFunc: schema.EnvDefaultFunc("CF_DEFAULT_QUOTA_NAME", "default"),
			},
			"app_logs_max": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of logs message which can be see when app creation is errored (-1 means all messages stored)",
				DefaultFunc: schema.EnvDefaultFunc("CF_APP_LOGS_MAX", 30),
			},
			"purge_when_delete": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_PURGE_WHEN_DELETE", false),
				Description: "Set to true to purge when deleting a resource (e.g.: service instance, service broker)",
			},
			"store_tokens_path": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_STORE_TOKENS_PATH", ""),
				Description: "Path to a file to store tokens used for login. (this is useful for sso, this avoid requiring each time sso passcode)",
			},
			"force_broker_not_fail_when_catalog_not_accessible": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_FORCE_BROKER_NOT_FAIL_CATALOG", false),
				Description: "Set to true to not trigger fail on catalog on service broker",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"cloudfoundry_info":                  dataSourceInfo(),
			"cloudfoundry_stack":                 dataSourceStack(),
			"cloudfoundry_stack_v3":              dataSourceStackV3(),
			"cloudfoundry_router_group":          dataSourceRouterGroup(),
			"cloudfoundry_user":                  dataSourceUser(),
			"cloudfoundry_domain":                dataSourceDomain(),
			"cloudfoundry_domain_v3": 	      dataSourceDomainV3(),
			"cloudfoundry_route":                 dataSourceRoute(),
			"cloudfoundry_asg":                   dataSourceAsg(),
			"cloudfoundry_org":                   dataSourceOrg(),
			"cloudfoundry_org_v3":                dataSourceOrgV3(),
			"cloudfoundry_org_quota":             dataSourceOrgQuota(),
			"cloudfoundry_space_quota":           dataSourceSpaceQuota(),
			"cloudfoundry_isolation_segment":     dataSourceIsolationSegment(),
			"cloudfoundry_space":                 dataSourceSpace(),
			"cloudfoundry_service_instance":      dataSourceServiceInstance(),
			"cloudfoundry_user_provided_service": dataSourceUserProvidedService(),
			"cloudfoundry_service_key":           dataSourceServiceKey(),
			"cloudfoundry_service":               dataSourceService(),
			"cloudfoundry_app":                   dataSourceApp(),
			"cloudfoundry_app_v3":                dataSourceAppV3(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"cloudfoundry_feature_flags":                 resourceConfig(),
			"cloudfoundry_user":                          resourceUser(),
			"cloudfoundry_domain":                        resourceDomain(),
			"cloudfoundry_private_domain_access":         resourcePrivateDomainAccess(),
			"cloudfoundry_asg":                           resourceAsg(),
			"cloudfoundry_org_quota":                     resourceOrgQuota(),
			"cloudfoundry_space_quota":                   resourceSpaceQuota(),
			"cloudfoundry_default_asg":                   resourceDefaultAsg(),
			"cloudfoundry_evg":                           resourceEvg(),
			"cloudfoundry_org":                           resourceOrg(),
			"cloudfoundry_space":                         resourceSpace(),
			"cloudfoundry_space_users":                   resourceSpaceUsers(),
			"cloudfoundry_org_users":                     resourceOrgUsers(),
			"cloudfoundry_service_broker":                resourceServiceBroker(),
			"cloudfoundry_service_plan_access":           resourceServicePlanAccess(),
			"cloudfoundry_service_instance":              resourceServiceInstance(),
			"cloudfoundry_service_key":                   resourceServiceKey(),
			"cloudfoundry_user_provided_service":         resourceUserProvidedService(),
			"cloudfoundry_buildpack":                     resourceBuildpack(),
			"cloudfoundry_route":                         resourceRoute(),
			"cloudfoundry_route_v3":                      resourceRouteV3(),
			"cloudfoundry_route_service_binding":         resourceRouteServiceBinding(),
			"cloudfoundry_app":                           resourceApp(),
			"cloudfoundry_isolation_segment":             resourceSegment(),
			"cloudfoundry_isolation_segment_entitlement": resourceSegmentEntitlement(),
			"cloudfoundry_network_policy":                resourceNetworkPolicy(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	c := managers.Config{
		Endpoint:                  strings.TrimSuffix(d.Get("api_url").(string), "/"),
		User:                      d.Get("user").(string),
		Password:                  d.Get("password").(string),
		SSOPasscode:               d.Get("sso_passcode").(string),
		CFClientID:                d.Get("cf_client_id").(string),
		CFClientSecret:            d.Get("cf_client_secret").(string),
		UaaClientID:               d.Get("uaa_client_id").(string),
		UaaClientSecret:           d.Get("uaa_client_secret").(string),
		SkipSslValidation:         d.Get("skip_ssl_validation").(bool),
		AppLogsMax:                d.Get("app_logs_max").(int),
		DefaultQuotaName:          d.Get("default_quota_name").(string),
		StoreTokensPath:           d.Get("store_tokens_path").(string),
		ForceNotFailBrokerCatalog: d.Get("force_broker_not_fail_when_catalog_not_accessible").(bool),
	}
	session, err := managers.NewSession(c)
	return session, diag.FromErr(err)
}
