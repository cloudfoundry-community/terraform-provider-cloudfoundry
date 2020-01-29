# Cloud Foundry Terraform Provider [![Build Status](https://travis-ci.org/cloudfoundry-community/terraform-provider-cf.svg?branch=master)](https://travis-ci.org/cloudfoundry-community/terraform-provider-cf)


Overview
--------

This Terraform provider plugin allows you to configure a Cloud Foundry environment declaratively using [HCL](https://github.com/hashicorp/hcl). 
The online documentation for the Terraform Cloud Foundry resource is available on the [wiki](https://github.com/cloudfoundry-community/terraform-provider-cf/wiki).

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.11.14
-	[Go](https://golang.org/doc/install) 1.12 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-cloudfoundry`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-cloudfoundry
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-cloudfoundry
$ make build
```

Using the provider
------------------

Download the release binary and copy it to the `$HOME/.terraform.d/plugins/<os>_<arch>/terraform-provider-cloudfoundry`. For example `/home/youruser/.terraform.d/plugins/linux_amd64/terraform-provider-cloudfoundry` for a Linux environment or `/Users/youruser/terraform.d/plugins/darwin_amd64/terraform-provider-cloudfoundry` for a MacOS environment (see https://www.terraform.io/docs/configuration/providers.html#third-party-plugins for more details).

Developing the Provider
-----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

Clone this repository to `GOPATH/src/github.com/terraform-providers/terraform-provider-cloudfoundry` as its packaging structure
has been defined such that it will be compatible with the Terraform provider plugin framwork in 0.10.x.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-cloudfoundry
...
```


Testing the Provider
--------------------

To test the provider you will need to run a local PCF Dev instance or launch it in AWS via the `scripts/pcfdev-up.sh`. Once the instance is running you will need to export the following environment variables.

```
export CF_API_URL=https://api.local.pcfdev.io
export CF_USER=admin
export CF_PASSWORD=admin
export CF_UAA_CLIENT_ID=admin
export CF_UAA_CLIENT_SECRET=admin-client-secret
export CF_CA_CERT=""
export CF_SKIP_SSL_VALIDATION=true
```

You can export the following environment variables to enable detail debug logs.

```
export CF_DEBUG=true
export CF_TRACE=debug.log
```

In order to run the tests locally, run.

```
cd cloudfoundry
TF_ACC=1 go test -v -timeout 120m .
```

To run the tests in AWS first launch PCFDev in AWS via `scripts/pcfdev-up.sh`, and then run.

```
make testacc
```

>> Acceptance tests are run against a PCF Dev instance in AWS before a release is created. Any other testing should be done using a local PCF Dev instance.

```sh
$ make testacc
```

Migration
---------

## 0.9.9 to 0.10.0

**Terraform version 0.11.14 at least is required**

Migration to 0.10.0 require you to change `cloudfoundry_app` and `cloudfoundry_buildpack` resource.
You must remove `url`, `git`, and `github_release` attributes from your resource and change to `path` according to 
the doc you can found here: https://github.com/cloudfoundry-community/terraform-provider-cf/wiki/resource_app#application-source--binary

Provider will migrate itself your tfstate and will download any non zip http(s) url in a folder `bits` in your current working directory.

For easier migration here the steps to follow with a tool which will do the change in your tf files directly without do anything:

1. *(Optional)* Migrate to terraform >= 0.12.x and follow migration step: https://www.terraform.io/upgrade-guides/0-12.html
2. Download `cf-hcl-migration` tool in the [release 0.10.0](https://github.com/cloudfoundry-community/terraform-provider-cf/releases/tag/v0.10.0).
this tool only change your tf files for preparing migration made by the provider.
3. Run `cf-hcl-migration ./` tool on the root folder of your terraform files.
4. Run `cf-hcl-migration <module path>` tool on the root folder of your terraform files.
**This tool migrate your terraform file config for only `cloudfoundry_app` and `cloudfoundry_buildpack` resource to change to new upload style
if your `tf` files doesn't use those resources no need to run this tool**
5. Run `terraform apply` in the root folder of your terraform.

## < 0.9.9 to 0.9.9

See the script in `scripts/migration` to migrate from versions below 0.9.9

Update doc
----------

You must update doc for resource and data sources in [website/docs](/website/docs).

For updating wiki please run `scripts/update-wiki.sh` this will create/update wiki files from [website/docs](/website/docs) and push it in the wiki.

Support
-------

You can reach us over [Slack](https://cloudfoundry.slack.com/messages/C7JRBR8CV/)

Terraform Links
---------------

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.svg)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)
