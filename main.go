package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: cloudfoundry.Provider,
	}
	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "registry.terraform.io/cloudfoundry-community/cloudfoundry"
	}
	plugin.Serve(opts)
}
