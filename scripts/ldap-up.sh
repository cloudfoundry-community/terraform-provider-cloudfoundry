#!/bin/sh

set -e

HOST_IP=$(ifconfig | awk '/inet.*[0-9]*\.[0-9]*\.[0-9]*\.[0-9]*/&& $2 != "127.0.0.1" { print $2 }' | head -1)

rm -fr .test_env/ldap
unzip .test_env/ldap.zip -d .test_env

docker run --name my-openldap-container --detach \
    -p 40389:389 \
    -v $(pwd)/.test_env/ldap/ldap:/var/lib/ldap \
    -v $(pwd)/.test_env/ldap/config:/etc/ldap/slapd.d \
    osixia/openldap:1.1.9

docker run --name my-phpldapadmin-container --detach \
    --env PHPLDAPADMIN_LDAP_HOSTS="#PYTHON2BASH:[{'$HOST_IP': [{'server': [{'port': 40389}]},{'login': [{'bind_id': 'cn=admin,dc=example,dc=org'}]}]}]" \
    -p 6443:443 \
    osixia/phpldapadmin:0.7.0

set +e