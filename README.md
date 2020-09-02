# Cloud Foundry Terraform Provider [![Build Status](https://travis-ci.org/cloudfoundry-community/terraform-provider-cloudfoundry.svg?branch=master)](https://travis-ci.org/cloudfoundry-community/terraform-provider-cloudfoundry)


Overview
--------

This Terraform provider plugin allows you to configure a Cloud Foundry environment declaratively using [HCL](https://github.com/hashicorp/hcl). 


The documentation is available at https://registry.terraform.io/providers/cloudfoundry-community/cloudfoundry

Using the provider
------------------

See doc at https://registry.terraform.io/providers/cloudfoundry-community/cloudfoundry , if you are under terraform 0.13 you can follow installation doc at https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/wiki  

Building The Provider
---------------------

Requirements:
- [Terraform](https://www.terraform.io/downloads.html) >= 0.11.14
- [Go](https://golang.org/doc/install)

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
Update doc
----------

You must update doc for resource and data sources in [docs](/docs), this will appears in the next release at https://registry.terraform.io/providers/cloudfoundry-community/cloudfoundry.

Support
-------

You can reach us over [Slack](https://cloudfoundry.slack.com/messages/C7JRBR8CV/)

Terraform Links
---------------

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.svg)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)
