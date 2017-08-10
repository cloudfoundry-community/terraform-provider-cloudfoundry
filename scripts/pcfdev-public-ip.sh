#!/bin/bash

function pcfdev_instance_detail() {
    aws ec2 describe-instances \
        | jq '.Reservations[] | .Instances[] | select(.State.Name != "terminated") | select(.Tags[] | .Value == "pcfdev")'
}

PCFDEV_INSTANCE_DETAIL=$(pcfdev_instance_detail)
if [[ -z $PCFDEV_INSTANCE_DETAIL ]]; then
    echo "ERROR! Unable to discover PCFDev instance."
    exit 1
fi

echo $(pcfdev_instance_detail | jq -r .PublicIpAddress)
