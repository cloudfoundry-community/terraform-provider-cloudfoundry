---
layout: "cloudfoundry"
page_title: "Provider: Cloud Foundry"
sidebar_current: "docs-cf-index"
description: |-
  The Cloud Foundry (cloudfoundry) provider is used to manage a Cloud Foundry environment. The provider needs to be configured with the proper credentials before it can be used.
---

# Cloud Foundry Provider

The Cloud Foundry (cloudfoundry) provider is used to interact with a
Cloud Foundry target to perform administrative configuration of platform 
resources, or user actions (such as pushing a cf app).

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Set the variable values in *.tfvars file
# or using -var="api_url=..." CLI option

variable "api_url" {}
variable "password" {}
variable "uaa_admin_client_secret" {}

# Configure the CloudFoundry Provider

provider "cloudfoundry" {
    api_url = var.api_url
    user = "admin"
    password = var.password
    uaa_client_id = "admin"
    uaa_client_secret = var.uaa_admin_client_secret
    skip_ssl_validation = true
    app_logs_max = 30
}
```

## Argument Reference

The following arguments are supported:

* `api_url` - (Required) API endpoint (e.g. https://api.local.pcfdev.io). This can also be specified
  with the `CF_API_URL` shell environment variable.

* `user` - (Optional) Cloud Foundry user. Defaults to "admin". This can also be specified
  with the `CF_USER` shell environment variable. Unless mentionned explicitly in a resource, CF admin permissions are not required.

* `password` - (Optional) Cloud Foundry user's password. This can also be specified
  with the `CF_PASSWORD` shell environment variable.

* `sso_passcode` - (Optional) A passcode provided by UAA single sign on. The equivalent of `cf login --sso-passcode`. This can also be specified
  with the `CF_SSO_PASSCODE` shell environment variable.

* `cf_client_id` - (Optional) The cf client ID to make request with a client instead of user. This can also be specified
  with the `CF_CLIENT_ID` shell environment variable.

* `cf_client_secret` - (Optional) The cf client secret to make request with a client instead of user. This can also be specified
  with the `CF_CLIENT_SECRET` shell environment variable.

* `uaa_client_id` - (Optional) The UAA admin client ID. Defaults to "admin". This can also be specified
  with the `CF_UAA_CLIENT_ID` shell environment variable.

* `uaa_client_secret` - (Optional) This secret of the UAA admin client. This can also be specified
  with the `CF_UAA_CLIENT_SECRET` shell environment variable.

* `skip_ssl_validation` - (Optional) Skip verification of the API endpoint - Not recommended!. Defaults to "false". This can also be specified
  with the `CF_SKIP_SSL_VALIDATION` shell environment variable.
  
* `default_quota_name` - (Optional, Default: `default`) Change the name of your default quota . This can also be specified
  with the `CF_DEFAULT_QUOTA_NAME` shell environment variable.
  
* `app_logs_max` - (Optional) Number of logs message which can be see when app creation is errored (-1 means all messages stored). Defaults to "30". This can also be specified
  with the `CF_APP_LOGS_MAX` shell environment variable.
  
* `purge_when_delete` - (Optional) Set to true to purge when deleting a resource (e.g.: service instance, service broker) . This can also be specified
  with the `CF_PURGE_WHEN_DELETE` shell environment variable.

* `store_tokens_path` - (Optional) Path to a file to store tokens used for login. (this is useful for sso, this avoid
  requiring each time sso passcode) . This can also be specified with the `CF_STORE_TOKENS_PATH` shell environment variable.
  
* `force_broker_not_fail_when_catalog_not_accessible` (Optional) Set to true to enforce `fail_when_catalog_not_accessible` to `true` to all broker for avoiding to be 
  stuck if broker has been deleted for example. This can also be specified with the `CF_FORCE_BROKER_NOT_FAIL_CATALOG` shell environment variable.
