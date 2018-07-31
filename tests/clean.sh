#!/bin/bash

echo "Start cleaning up potentially leaking resources from previous test executions. Warnings about missing resources should be ignored"

CF_SPACE=pcfdev-space
CF_ORG=pcfdev-org

set -e # Exit if the login fails (not set or wrongly set!)
cf api $CF_API_URL --skip-ssl-validation
cf login -u $CF_USER -p $CF_PASSWORD -o $CF_ORG -s $CF_SPACE
set +e

# Please add any further resources do not get destroyed

# Delete apps

cf delete -f php-app
cf delete -f basic-auth-router
cf delete -f basic-auth-broker
cf delete -f fake-service-broker
cf delete -f test-app
cf delete -f test-docker-app

# Delete org and security gorups

cf delete-org -f organization-one
cf delete-org -f myorg
cf delete-security-group -f app-services1
cf delete-security-group -f app-services2
cf delete-security-group -f app-services3
cf delete-security-group -f app-services

# Delete services and service instances

cf delete-service -f basic-auth
cf delete-service -f rabbitmq
cf purge-service-offering -f p-basic-auth
cf delete-service-broker -f basic-auth

# Delete routes

cf delete-route -f $CF_TEST_APP_DOMAIN --hostname php-app
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname php-app-other
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname basic-auth-router
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname basic-auth-broker
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname test-app
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname test-docker-app
cf delete-route -f $CF_TEST_APP_DOMAIN --hostname fake-service-broker
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

# Sanity checks

CF_SPACE_GUID=`cf space --guid $CF_SPACE`
CF_ORG_GUID=`cf org --guid $CF_ORG`

if [ `cf curl "/v2/apps?q=space_guid:$CF_SPACE_GUID" | jq ".total_results"` -ne "0" ]; then
   echo "ERROR: The acceptance environment contains some residual apps, run \"cf a\" - please clean them up";
   exit 1;
fi

if [ `cf curl "/v2/routes?q=organization_guid:$CF_ORG_GUID" | jq ".total_results"` -ne "0" ]; then
   echo "ERROR: The acceptance environment contains some residual routes, run \"cf routes\" - please clean them up";
   exit 1;
fi

if [ `cf curl "/v2/service_instances?q=organization_guid:$CF_ORG_GUID" | jq ".total_results"` -ne "0" ]; then
   echo "ERROR: The acceptance environment contains some residual service instances, run \"cf s\" - please clean them up";
   exit 1;
fi

echo "Completed cleaning up potentially leaking resources from previous test executions."
exit 0