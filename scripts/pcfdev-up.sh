#!/bin/bash

set -e

SECONDS=0
INGRESS_ALLOWED_IP=$(curl -s http://whatismyip.akamai.com/)

function pcfdev_instance_detail() {
    aws ec2 describe-instances \
        | jq '.Reservations[] | .Instances[] | select(.State.Name != "terminated") | select(.Tags[] | .Value == "pcfdev")'
}

# Deploy PCFDev instance if not already deployed

mkdir -p .test_env
pushd .test_env

PCFDEV_INSTANCE_DETAIL=$(pcfdev_instance_detail)
if [[ -n $PCFDEV_INSTANCE_DETAIL ]]; then

    INSTANCE_ID=$(echo $PCFDEV_INSTANCE_DETAIL | jq -r .InstanceId)

    echo "Terminating existing PCFDev instance having ID $ID..."
    aws ec2 terminate-instances --instance-id $INSTANCE_ID >/dev/null

    PCFDEV_INSTANCE_DETAIL=$(pcfdev_instance_detail)
    INSTANCE_STATE=$(echo $PCFDEV_INSTANCE_DETAIL | jq -r .State.Name)

    i=0
    while [[ $i -lt 120 && -n $INSTANCE_STATE && $INSTANCE_STATE != terminated ]]; do

        echo "Waiting for PCFDev instance to terminate..."
        sleep 5
        i=$((i+1))

        PCFDEV_INSTANCE_DETAIL=$(pcfdev_instance_detail)
        INSTANCE_STATE=$(echo $PCFDEV_INSTANCE_DETAIL | jq -r .State.Name)
    done
    if [[ $i -eq 120 ]]; then
        echo "ERROR! Timed out waiting for PCFDev instance to terminate."
        exit 1
    fi

    rm -fr .vagrant
fi

echo "Creating new PCFDev instance..."

# Download PCFDev for AWS

case $(uname) in
    Linux)
        wget -O pivnet https://github.com/pivotal-cf/pivnet-cli/releases/download/v0.0.49/pivnet-linux-amd64-0.0.49
        ;;
    Darwin)
        wget -O pivnet https://github.com/pivotal-cf/pivnet-cli/releases/download/v0.0.49/pivnet-darwin-amd64-0.0.49
        ;;
    *)
        echo "ERROR: Unable download 'pivnet' CLI to download PCFDev."
        exit 1
esac

chmod 0755 ./pivnet

./pivnet login --api-token=$PIVNET_TOKEN
./pivnet accept-eula --product-slug pcfdev --release-version aws

PCFDEV_PRODUCT_DETAIL=$(./pivnet product-files --product-slug=pcfdev --release-version aws --format json)
PCFDEV_DOWNLOAD_URL=$(echo $PCFDEV_PRODUCT_DETAIL \
    | jq -r ".[] | select(.file_version == \"v0.26.0 for PCF 1.10.0\") | ._links.download.href")
PCFDEV_FILE_NAME=$(echo $PCFDEV_PRODUCT_DETAIL \
    | jq -r ".[] | select(.file_version == \"v0.26.0 for PCF 1.10.0\") | .name")

wget --post-data '' --header "Authorization: Token $PIVNET_TOKEN" -O $PCFDEV_FILE_NAME $PCFDEV_DOWNLOAD_URL
unzip -o $PCFDEV_FILE_NAME

PCFDEV_UNZIP_DIR=${PCFDEV_FILE_NAME%%.zip*}
vagrant box add pcfdev/pcfdev $PCFDEV_UNZIP_DIR/aws-*.box -f

# Create EC2 security to enable access only from this client

VPC_ID=$(aws ec2 describe-vpcs \
    | jq -r '.Vpcs[] | select(.IsDefault == true) | .VpcId')

[[ -z `aws ec2 describe-security-groups | jq -r '.SecurityGroups[] | select(.GroupName == "pcfdev") | .GroupName'` ]] ||
    aws ec2 delete-security-group --group-name pcfdev >/dev/null

aws ec2 create-security-group --group-name pcfdev --description "PCFDev secure ingress" --vpc-id $VPC_ID >/dev/null
aws ec2 authorize-security-group-ingress --group-name pcfdev --protocol tcp --port 80 --cidr $INGRESS_ALLOWED_IP/32
aws ec2 authorize-security-group-ingress --group-name pcfdev --protocol tcp --port 443 --cidr $INGRESS_ALLOWED_IP/32
aws ec2 authorize-security-group-ingress --group-name pcfdev --protocol tcp --port 22 --cidr $INGRESS_ALLOWED_IP/32

# Create key-pair for PCFDev instance

[[ -z `aws ec2 describe-key-pairs | jq -r '.KeyPairs[] | select(.KeyName == "pcfdev") | .KeyName'` ]] ||
    aws ec2 delete-key-pair --key-name pcfdev >/dev/null

rm -f pcfdev.pem pcfdev.pem.pub
ssh-keygen -t rsa -b 4096 -N "" -f pcfdev.pem
aws ec2 import-key-pair --key-name pcfdev --public-key-material "$(cat pcfdev.pem.pub)" >/dev/null

# Launch PCFDev instance using Vagrant

cat <<VAGRANTFILE > Vagrantfile
Vagrant.configure("2") do |config|
  config.vm.box = "pcfdev/pcfdev"
  config.vm.box_version = "0"

  config.vm.provider "aws" do |aws, override|    
    aws.region = ENV["AWS_DEFAULT_REGION"]
    aws.access_key_id = ENV["AWS_ACCESS_KEY_ID"]
    aws.secret_access_key = ENV["AWS_SECRET_ACCESS_KEY"]
    aws.keypair_name = "pcfdev"
    aws.security_groups = ["pcfdev"]
    aws.associate_public_ip = false
    override.ssh.private_key_path = "./pcfdev.pem"
  end
end
VAGRANTFILE

set -x

vagrant up --provider=aws --no-provision

PCFDEV_INSTANCE_DETAIL=$(pcfdev_instance_detail)
if [[ -z $PCFDEV_INSTANCE_DETAIL ]]; then
    echo "ERROR! Unable to discover PCFDev instance which was just launched."
    exit 1
fi

INSTANCE_ID=$(echo $PCFDEV_INSTANCE_DETAIL | jq -r .InstanceId)
INSTANCE_PUBLIC_IP=$(echo $PCFDEV_INSTANCE_DETAIL | jq -r .PublicIpAddress)

echo "PCFDev instance's ID is: $INSTANCE_ID"
echo "Provisioning PCFDev on $INSTANCE_PUBLIC_IP.xip.io domain for access from $INGRESS_ALLOWED_IP..."

export PCFDEV_DOMAIN=$INSTANCE_PUBLIC_IP.xip.io
vagrant provision

duration=$SECONDS
echo "Time to provision PCDDev via Vagrant: $(($duration / 60)) minutes and $(($duration % 60)) seconds."

popd

set +x
set +e
