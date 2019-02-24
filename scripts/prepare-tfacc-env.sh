#!/bin/bash

set -xe

which om >/dev/null 2>&1
if [[ $? -ne 0 ]]; then
    echo "The OM CLI is not installed. Please download it from https://github.com/pivotal-cf/om/releases."
    exit 1
fi

root_dir=$(cd $(dirname $BASH_SOURCE)/.. && pwd)

opsman_endpoint=$OPSMAN_ENDPOINT
opsman_user=$OPSMAN_USER
opsman_password=$OPSMAN_PASSWORD

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

cat <<EOF > $root_dir/.tfacc_env
#!/bin/bash

export CF_API_URL=https://api.$cf_sys_domain
export CF_USER=$cf_user
export CF_PASSWORD=$cf_password
export CF_UAA_CLIENT_ID=$cf_uaa_client_id
export CF_UAA_CLIENT_SECRET=$cf_uaa_client_secret
export CF_CA_CERT=""
export CF_SKIP_SSL_VALIDATION=true

export TEST_APP_DOMAIN=$cf_apps_domain
export TEST_DEFAULT_ASG=$TEST_DEFAULT_ASG
export TEST_ORG_NAME=$TEST_ORG_NAME
export TEST_SPACE_NAME=$TEST_SPACE_NAME

export TEST_SERVICE_BROKER_URL=$TEST_SERVICE_BROKER_URL
export TEST_SERVICE_BROKER_USER=$TEST_SERVICE_BROKER_USER
export TEST_SERVICE_BROKER_PASSWORD=$TEST_SERVICE_BROKER_PASSWORD
export TEST_SERVICE_PLAN_PATH=$TEST_SERVICE_PLAN_PATH

export TEST_SERVICE_1=$TEST_SERVICE_1
export TEST_SERVICE_2=$TEST_SERVICE_2
export TEST_SERVICE_PLAN=$TEST_SERVICE_PLAN
EOF

chmod +x $root_dir/.tfacc_env

if [[ -n $GITHUB_TOKEN ]]; then

    if [[ "$(uname)" == "Darwin" ]]; then
        sed -i '' '/  - secure:.*/d' $root_dir/.travis.yml
    else
        sed -i '/  - secure:.*/d' $root_dir/.travis.yml
    fi

    travis login --github-token $GITHUB_TOKEN
    travis encrypt CF_API_URL=https://api.$cf_sys_domain --add
    travis encrypt CF_USER=$cf_user --add
    travis encrypt CF_PASSWORD=$cf_password --add
    travis encrypt CF_UAA_CLIENT_ID=$cf_uaa_client_id --add
    travis encrypt CF_UAA_CLIENT_SECRET=$cf_uaa_client_secret --add
    travis encrypt CF_CA_CERT="" --add
    travis encrypt CF_SKIP_SSL_VALIDATION=true --add

    travis encrypt TEST_APP_DOMAIN=$cf_apps_domain --add
    travis encrypt TEST_DEFAULT_ASG=$TEST_DEFAULT_ASG --add
    travis encrypt TEST_ORG_NAME=$TEST_ORG_NAME --add
    travis encrypt TEST_SPACE_NAME=$TEST_SPACE_NAME --add

    travis encrypt TEST_SERVICE_BROKER_URL=$TEST_SERVICE_BROKER_URL --add
    travis encrypt TEST_SERVICE_BROKER_USER=$TEST_SERVICE_BROKER_USER --add
    travis encrypt TEST_SERVICE_BROKER_PASSWORD=$TEST_SERVICE_BROKER_PASSWORD --add
    travis encrypt TEST_SERVICE_PLAN_PATH=$TEST_SERVICE_PLAN_PATH --add

    travis encrypt TEST_SERVICE_1=$TEST_SERVICE_1 --add
    travis encrypt TEST_SERVICE_2=$TEST_SERVICE_2 --add
    travis encrypt TEST_SERVICE_PLAN=$TEST_SERVICE_PLAN --add
fi

set +xe
