package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cloudfoundry.Provider})

}
