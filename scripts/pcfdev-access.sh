#!/bin/bash

set -e

TAG="$(git tag -l --points-at HEAD)"
if [[ -z "$TAG" ]] ; then
    echo "Git commit does not have a release tag so acceptance tests will not run."
    touch no_acc
    exit 0
fi

INGRESS_ALLOWED_IP=$(curl -s http://whatismyip.akamai.com/)

function pcfdev_instance_detail() {
    aws ec2 describe-instances \
        | jq '.Reservations[] | .Instances[] | select(.State.Name != "terminated") | select(.Tags[] | .Value == "pcfdev")'
}

PCFDEV_INSTANCE_DETAIL=$(pcfdev_instance_detail)
if [[ -z $PCFDEV_INSTANCE_DETAIL ]]; then
    echo "ERROR! Unable to discover PCFDev instance."
    exit 1
fi

INSTANCE_STATE=$(echo $PCFDEV_INSTANCE_DETAIL | jq -r .State.Name)
if [[ $INSTANCE_STATE != running ]]; then
    echo "ERROR! PCFDev instance is not running."
    exit 1
fi

# Setup EC2 security to enable access only from this client

INGRESS_ALLOWED_CIDR=$(aws ec2 describe-security-groups --group-name pcfdev \
    | jq -r '.SecurityGroups[] | .IpPermissions[] | .IpRanges[] | .CidrIp' | uniq | head -1)

if [[ ${INGRESS_ALLOWED_CIDR%/*} != $INGRESS_ALLOWED_IP ]]; then

    echo "Updating security to group to allow ingress from $INGRESS_ALLOWED_IP..."

    if [[ -n $INGRESS_ALLOWED_CIDR ]]; then
        aws ec2 revoke-security-group-ingress --group-name pcfdev --protocol tcp --port 80 --cidr $INGRESS_ALLOWED_CIDR
        aws ec2 revoke-security-group-ingress --group-name pcfdev --protocol tcp --port 443 --cidr $INGRESS_ALLOWED_CIDR
        aws ec2 revoke-security-group-ingress --group-name pcfdev --protocol tcp --port 22 --cidr $INGRESS_ALLOWED_CIDR
    fi

    aws ec2 authorize-security-group-ingress --group-name pcfdev --protocol tcp --port 80 --cidr $INGRESS_ALLOWED_IP/32
    aws ec2 authorize-security-group-ingress --group-name pcfdev --protocol tcp --port 443 --cidr $INGRESS_ALLOWED_IP/32
    aws ec2 authorize-security-group-ingress --group-name pcfdev --protocol tcp --port 22 --cidr $INGRESS_ALLOWED_IP/32
fi

set +e