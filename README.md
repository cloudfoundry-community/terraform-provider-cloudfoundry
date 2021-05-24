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

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.12+ is *required*). 

1. git clone this repo
2. simply run `go build .` for building the provider
3. add a file at `${HOME}/.terraformrc` and set this content
```hcl
providers {
	cloudfoundry = "path/where/you/have/clone/repo/terraform-provider-cloudfoundry"
}
```

That's override the path where to found provider binary to use your development version. 


Debugging the Provider
----------------------

You can build a binary of the provider and then starting it with the `-debug` flag. Example

```shell
$ dlv exec --headless ./terraform-provider-cloudfoundry -- -debug
```

Connect your debugger (whether it's your IDE or the debugger client) to the debugger server. Have it continue execution (it pauses the process by default) and it will print output like the following to `stdout`:

```text
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:

        TF_REATTACH_PROVIDERS='{"registry.terraform.io/cloudfoundry-community/cloudfoundry":{"Protocol":"grpc","Pid":3382870,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin713096927"}}}'
```

Running Terraform with the provider in Debug Mode
-------------------------------------------------
Copy the line starting with `TF_REATTACH_PROVIDERS` from your provider's output. Either export it, or prefix every Terraform command with it:

```shell
TF_REATTACH_PROVIDERS='{"registry.terraform.io/cloudfoundry-community/cloudfoundry":{"Protocol":"grpc","Pid":3382870,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin713096927"}}}' terraform apply
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
