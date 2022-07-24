# Cloud Foundry Terraform Provider [![Build Status](https://travis-ci.org/cloudfoundry-community/terraform-provider-cloudfoundry.svg?branch=master)](https://travis-ci.org/cloudfoundry-community/terraform-provider-cloudfoundry)

# Specific to v3 migration
You don't need to do any setup except for installing the latest version of go. I'm using vscode with go extension installed.

## API docs
It is very important to read both v3 and v2 api docs to see the differences between the two.
- v3 api docs: http://v3-apidocs.cloudfoundry.org/version/3.122.0/
- v2 api docs: https://apidocs.cloudfoundry.org/16.22.0/

## Override with ~/.terraformrc
If you don't have anything in ~/.terraformrc you can simply run `make local-install` to build, install and configure dev_override for the provider
```shell
make local-install
```
After installing the provider, you can run terraform scripts using the "cloudfoundry" provider directly, c.f examples/main.tf for more information
```HCL
provider "cloudfoundry" {
    api_url              = "https://api.cf.sap.hana.ondemand.com" // eu10-canary
    sso_passcode         = var.sso_passcode
    store_tokens_path    = "tokens.txt"
}
```

**! It's important that you run `make clean` after your tests to remove the configs in ~/.terraformrc **

## No overriding ~/.terraformrc
If you have some configs in ~/.terraformrc that you don't want to lose, you can run `make local-build-only` and manually configure ~/.terraformrc by adding this dev_override
```config
provider_installation {
  dev_overrides {
    "cloudfoundry" = "~/.terraform.d/plugins/linux_amd64/terraform-provider-cloudfoundry"
  }

  # all the other providers, install them as usual
  direct {}
}
```
After you can run tf scripts as mentioned above.

Original docs
-------------
Original README for the provider is [here](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry)
