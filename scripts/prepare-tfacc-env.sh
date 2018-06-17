#!/bin/bash

set -xe

which om >/dev/null 2>&1
if [[ $? -ne 0 ]]; then
    echo "The OM CLI is not installed. Please download it from https://github.com/pivotal-cf/om/releases."
    exit 1
fi

root_dir=$(cd $(dirname $BASH_SOURCE)/.. && pwd)

opsman_endpoint=${1:-$OPSMAN_ENDPOINT}
opsman_user=${2:-$OPSMAN_USER}
opsman_password=${3:-$OPSMAN_PASSWORD}

if [[ -z $opsman_endpoint && -z $opsman_user && -z opsman_password ]]; then
    echo "USAGE: ./prepare-tfacc-env.sh $OPSMAN_ENDPOINT $OPSMAN_USER $OPSMAN_PASSWORD"
    exit 1
fi

om_cli="om --skip-ssl-validation --target https://$opsman_endpoint --username $opsman_user --password $opsman_password"

$om_cli curl -p /api/installation_settings > /tmp/installation_settings.json
cf_sys_domain=$(cat /tmp/installation_settings.json \
    | jq -r '.products[] | select(.installation_name | match("cf-.*")) | .jobs[] | select(.installation_name == "cloud_controller") | .properties[] | select(.identifier == "system_domain") | .value')
cf_apps_domain=$(cat /tmp/installation_settings.json \
    | jq -r '.products[] | select(.installation_name | match("cf-.*")) | .jobs[] | select(.installation_name == "cloud_controller") | .properties[] | select(.identifier == "apps_domain") | .value')

cf_user=$($om_cli credentials -p cf -c .uaa.admin_credentials -f identity)
cf_password=$($om_cli credentials -p cf -c .uaa.admin_credentials -f password)
cf_uaa_client_id=$($om_cli credentials -p cf -c .uaa.admin_client_credentials -f identity)
cf_uaa_client_secret=$($om_cli credentials -p cf -c .uaa.admin_client_credentials -f password)
cf_test_redis_broker_user=$($om_cli credentials -p p-redis -c .cf-redis-broker.broker_http_auth_credentials -f identity)
cf_test_redis_broker_password=$($om_cli credentials -p p-redis -c .cf-redis-broker.broker_http_auth_credentials -f password)

cat <<EOF > $root_dir/.tfacc_env
#!/bin/bash

export CF_API_URL=https://api.$cf_sys_domain
export CF_USER=$cf_user
export CF_PASSWORD=$cf_password
export CF_UAA_CLIENT_ID=$cf_uaa_client_id
export CF_UAA_CLIENT_SECRET=$cf_uaa_client_secret
export CF_CA_CERT=""
export CF_SKIP_SSL_VALIDATION=true

export CF_TEST_APP_DOMAIN=$cf_apps_domain

export CF_TEST_REDIS_BROKER_USER=$cf_test_redis_broker_user
export CF_TEST_REDIS_BROKER_PASSWORD=$cf_test_redis_broker_password

export CF_TEST_DEFAULT_ASG=default_security_group
EOF

if [[ -n $GITHUB_TOKEN ]]; then
    travis login --github-token $GITHUB_TOKEN
    travis encrypt CF_API_URL=https://api.$cf_sys_domain --add --override
    travis encrypt CF_USER=$cf_user --add
    travis encrypt CF_PASSWORD=$cf_password --add
    travis encrypt CF_UAA_CLIENT_ID=$cf_uaa_client_id --add
    travis encrypt CF_UAA_CLIENT_SECRET=$cf_uaa_client_secret --add
    travis encrypt CF_CA_CERT="" --add
    travis encrypt CF_SKIP_SSL_VALIDATION=true --add
    travis encrypt CF_TEST_APP_DOMAIN=$cf_apps_domain --add
    travis encrypt CF_TEST_REDIS_BROKER_USER=$cf_test_redis_broker_user --add
    travis encrypt CF_TEST_REDIS_BROKER_PASSWORD=$cf_test_redis_broker_password --add
    travis encrypt CF_TEST_DEFAULT_ASG=default_security_group --add
fi

cf login --skip-ssl-validation -a https://api.$cf_sys_domain -u $cf_user -p $cf_password -o system -s system
cf create-org pcfdev-org >/dev/null 2>&1
cf target -o pcfdev-org >/dev/null 2>&1
cf create-space pcfdev-space -o pcfdev-org >/dev/null 2>&1
cf create-shared-domain tcp.apps.pcf.cf1.tfacc.pcfs.io --router-group default-tcp >/dev/null 2>&1
cf enable-feature-flag diego_docker >/dev/null 2>&1

set +xe
