package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceInfo() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceInfoRead,

		Schema: map[string]*schema.Schema{
			"api_version": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"uaa_endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"routing_endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging_endpoint": &schema.Schema{
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "Not exists anymore in new cloud foundry",
			},
			"doppler_endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceInfoRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	info := session.ClientV3.Info
	infoV2, _, err := session.ClientV2.Info()
	if err != nil {
		return err
	}
	d.Set("api_version", info.CloudControllerAPIVersion())
	d.Set("auth_endpoint", infoV2.AuthorizationEndpoint)
	d.Set("uaa_endpoint", info.UAA())
	d.Set("routing_endpoint", info.Routing())
	d.Set("logging_endpoint", strings.Replace(info.Logging(), "doppler", "loggregator", 1))
	d.Set("doppler_endpoint", info.Logging())

	d.SetId("info")
	return nil
}
