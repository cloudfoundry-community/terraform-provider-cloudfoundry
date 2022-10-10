# Cloud Foundry Terraform Provider [![Build Status](https://travis-ci.org/cloudfoundry-community/terraform-provider-cloudfoundry.svg?branch=master)](https://travis-ci.org/cloudfoundry-community/terraform-provider-cloudfoundry)

# Using the internal release
Please refer to the [How to use section](../../wiki/How-to-use-the-provider) of the Wiki for a detailed guide

# Specific to v3 migration
You don't need to do any setup except for installing the latest version of go. I'm using vscode with go extension installed.

## API docs
It is very important to read both v3 and v2 api docs to see the differences between the two.
- v3 api docs: http://v3-apidocs.cloudfoundry.org/version/3.122.0/
- v2 api docs: https://apidocs.cloudfoundry.org/16.22.0/

## Common issues when pulling dependencies from github.wdf.sap.corp
There are many ways to go about pulling dependencies from a private git repository, below is my preferred way that has been proven to work

1. Add configuration to your git config at ~/.gitconfig. This will override https requests by ssh requests. You can follow the tutorial on github on how to add your ssh key to your github profile : https://docs.github.com/enterprise/3.5/articles/generating-an-ssh-key/
```sh
# Add this to the end of your gitconfig file at ~/.gitconfig
[url "ssh://git@github.wdf.sap.corp/"]
        insteadOf = https://github.wdf.sap.corp/
```

2. Add SAP's rootca certificate locally so that the machine validates the certificate coming from github.wdf.sap.corp
```sh
# Download CA certificate via command: 
wget http://aia.pki.co.sap.com/aia/SAP%20Global%20Root%20CA.crt

# Set up it on Linux via https://manuals.gfi.com/en/kerio/connect/content/server-configuration/ssl-certificates/adding-trusted-root-certificates-to-the-server-1605.html

# Copy your CA to dir /usr/local/share/ca-certificates/
sudo cp foo.crt /usr/local/share/ca-certificates/foo.crt
# Update the CA store: 
sudo update-ca-certificates
```


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
