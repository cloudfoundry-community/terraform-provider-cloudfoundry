package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cloudfoundry.Provider})

}
