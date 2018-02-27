#!/bin/bash

cf target -o pcfdev-org -s pcfdev-space


cf delete-service -f basic-auth
cf delete -f php-app
cf delete -f basic-auth-app
cf delete -f basic-auth-broker
cf delete-route -f local.pcfdev.io --hostname php-app
cf delete-route -f local.pcfdev.io --hostname basic-auth-app
cf delete-route -f local.pcfdev.io --hostname basic-auth-broker

url=$(cf curl /v2/service_brokers | jq -r '.resources[] | select(.entity.name | contains("basic-auth")) | .metadata.url')
if [ ! -z "${url}" ]; then
    echo deleting ${url}
    cf curl -X DELETE ${url}
fi

