#!/bin/bash

cf target -o pcfdev-org -s pcfdev-space

cf unbind-route-service -f local.pcfdev.io  basic-auth --hostname php-app
cf delete-service -f basic-auth
cf delete -f php-app
cf delete -f basic-auth-router
cf delete -f basic-auth-broker
cf delete-route -f local.pcfdev.io --hostname php-app
cf delete-route -f local.pcfdev.io --hostname php-app-other
cf delete-route -f local.pcfdev.io --hostname basic-auth-router
cf delete-route -f local.pcfdev.io --hostname basic-auth-broker
cf purge-service-offering -f p-basic-auth
cf delete-service-broker -f basic-auth

# url=$(cf curl /v2/service_brokers | jq -r '.resources[] | select(.entity.name | contains("basic-auth")) | .metadata.url')
# if [ ! -z "${url}" ]; then
#     echo deleting ${url}
#     cf curl -X DELETE ${url}
# fi

