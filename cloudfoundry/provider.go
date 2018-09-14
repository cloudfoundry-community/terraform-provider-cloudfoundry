package cloudfoundry

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider -
func Provider() terraform.ResourceProvider {

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_API_URL", ""),
			},
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_USER", ""),
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_PASSWORD", ""),
			},
			"uaa_client_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_UAA_CLIENT_ID", ""),
			},
			"uaa_client_secret": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_UAA_CLIENT_SECRET", ""),
			},
			"ca_cert": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_CA_CERT", ""),
			},
			"skip_ssl_validation": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CF_SKIP_SSL_VALIDATION", "true"),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"cloudfoundry_info":         dataSourceInfo(),
			"cloudfoundry_stack":        dataSourceStack(),
			"cloudfoundry_router_group": dataSourceRouterGroup(),
			"cloudfoundry_user":         dataSourceUser(),
			"cloudfoundry_domain":       dataSourceDomain(),
			"cloudfoundry_asg":          dataSourceAsg(),
			"cloudfoundry_org":          dataSourceOrg(),
			"cloudfoundry_org_quota":    dataSourceOrgQuota(),
			"cloudfoundry_space_quota":  dataSourceSpaceQuota(),
			"cloudfoundry_space":        dataSourceSpace(),
			"cloudfoundry_service":      dataSourceService(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"cloudfoundry_feature_flags":         resourceConfig(),
			"cloudfoundry_user":                  resourceUser(),
			"cloudfoundry_domain":                resourceDomain(),
			"cloudfoundry_private_domain_access": resourcePrivateDomainAccess(),
			"cloudfoundry_asg":                   resourceAsg(),
			"cloudfoundry_org_quota":             resourceOrgQuota(),
			"cloudfoundry_space_quota":           resourceSpaceQuota(),
			"cloudfoundry_default_asg":           resourceDefaultAsg(),
			"cloudfoundry_evg":                   resourceEvg(),
			"cloudfoundry_org":                   resourceOrg(),
			"cloudfoundry_space":                 resourceSpace(),
			"cloudfoundry_service_broker":        resourceServiceBroker(),
			"cloudfoundry_service_plan_access":   resourceServicePlanAccess(),
			"cloudfoundry_service_instance":      resourceServiceInstance(),
			"cloudfoundry_service_key":           resourceServiceKey(),
			"cloudfoundry_user_provided_service": resourceUserProvidedService(),
			"cloudfoundry_buildpack":             resourceBuildpack(),
			"cloudfoundry_route":                 resourceRoute(),
			"cloudfoundry_route_service_binding": resourceRouteServiceBinding(),
			"cloudfoundry_app":                   resourceApp(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	config := Config{
		endpoint:          d.Get("api_url").(string),
		User:              d.Get("user").(string),
		Password:          d.Get("password").(string),
		UaaClientID:       d.Get("uaa_client_id").(string),
		UaaClientSecret:   d.Get("uaa_client_secret").(string),
		CACert:            d.Get("ca_cert").(string),
		SkipSslValidation: d.Get("skip_ssl_validation").(bool),
	}
	return config.Client()
}
