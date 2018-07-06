#!/bin/bash

cf login -u $CF_USER -p $CF_PASSWORD -o pcfdev-org -s pcfdev-space

# Delete apps
cf delete -f php-app
cf delete -f basic-auth-router
cf delete -f basic-auth-broker
cf delete -f test-app
cf delete -f test-docker-app

# Delete org and security gorups

cf delete-org -f organization-one
cf delete-security-group -f app-services1
cf delete-security-group -f app-services2
cf delete-security-group -f app-services3
cf delete-security-group -f app-services

# Delete services and service instances

cf delete-service -f basic-auth
cf purge-service-offering -f p-basic-auth
cf delete-service-broker -f basic-auth

# Delete routes

cf delete-route -f $CF_TEST_APP_DOMAIN --hostname php-app
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname php-app-other
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname basic-auth-router
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname basic-auth-broker
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname test-app
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname test-docker-app
cf unbind-route-service -f $CF_TEST_APP_DOMAIN basic-auth --hostname php-app

# Delete users
cf delete-user manager1@acme.com -f
cf delete-user auditor@acme.com -f
cf delete-user teamlead@acme.com -f
cf delete-user developer1@acme.com -f
cf delete-user developer2@acme.com -f
cf delete-user developer3@acme.com -f
cf delete-user cf-admin -f

# url=$(cf curl /v2/service_brokers | jq -r '.resources[] | select(.entity.name | contains("basic-auth")) | .metadata.url')
# if [ ! -z "${url}" ]; then
#     echo deleting ${url}
#     cf curl -X DELETE ${url}
# fi

