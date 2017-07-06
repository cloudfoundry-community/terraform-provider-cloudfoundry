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
			"cf_info":         dataSourceInfo(),
			"cf_stack":        dataSourceStack(),
			"cf_router_group": dataSourceRouterGroup(),
			"cf_user":         dataSourceUser(),
			"cf_domain":       dataSourceDomain(),
			"cf_asg":          dataSourceAsg(),
			"cf_quota":        dataSourceQuota(),
			"cf_org":          dataSourceOrg(),
			"cf_space":        dataSourceSpace(),
			"cf_service":      dataSourceService(),
			"cf_service_plan": dataSourceServicePlan(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"cf_config":                resourceConfig(),
			"cf_user":                  resourceUser(),
			"cf_user_org_role":         resourceUserOrgRole(),
			"cf_domain":                resourceDomain(),
			"cf_quota":                 resourceQuota(),
			"cf_asg":                   resourceAsg(),
			"cf_default_asg":           resourceDefaultAsg(),
			"cf_evg":                   resourceEvg(),
			"cf_org":                   resourceOrg(),
			"cf_space":                 resourceSpace(),
			"cf_service_broker":        resourceServiceBroker(),
			"cf_service_access":        resourceServiceAccess(),
			"cf_service_instance":      resourceServiceInstance(),
			"cf_user_provided_service": resourceUserProvidedService(),
			"cf_buildpack":             resourceBuildpack(),
			"cf_route":                 resourceRoute(),
			"cf_app":                   resourceApp(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	if err := initRepoManager(); err != nil {
		return nil, err
	}

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
