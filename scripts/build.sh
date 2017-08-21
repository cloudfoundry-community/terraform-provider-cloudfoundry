#!/bin/bash

if [[ -z "$TRAVIS_TAG" ]] ; then
    echo "Git commit does not have a release tag so acceptance tests will not run."
    
    make build
    exit 0
fi

scripts/pcfdev-up.sh
trap scripts/pcfdev-destroy.sh EXIT

function pcfdev_instance_detail() {
    aws ec2 describe-instances \
        | jq '.Reservations[] | .Instances[] | select(.State.Name != "terminated") | select(.Tags[] | .Value == "pcfdev")'
}

set -e

PCFDEV_INSTANCE_DETAIL=$(pcfdev_instance_detail)
if [[ -z $PCFDEV_INSTANCE_DETAIL ]]; then
    echo "ERROR! Unable to discover PCFDev instance."
    exit 1
fi

PUBLIC_IP=$(echo $PCFDEV_INSTANCE_DETAIL | jq -r .PublicIpAddress)
PRIVATE_IP=$(echo $PCFDEV_INSTANCE_DETAIL | jq -r .PrivateIpAddress)

ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i .test_env/pcfdev.pem ubuntu@$PUBLIC_IP <<EOF
#!/bin/bash

rm -fr /tmp/gopath
mkdir -p /tmp/gopath/src/github.com/terraform-providers/terraform-provider-cloudfoundry
EOF

scp -q -r -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i .test_env/pcfdev.pem \
    ./* ubuntu@$PUBLIC_IP:/tmp/gopath/src/github.com/terraform-providers/terraform-provider-cloudfoundry

ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i .test_env/pcfdev.pem ubuntu@$PUBLIC_IP <<EOF
#!/bin/bash

go version | grep go1.8 >/dev/null 2>&1
[[ \$? -eq 0 ]] || \
    sudo rm -fr /usr/local/go/ && curl https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz | sudo tar xz -C /usr/local

set -ex

export GOROOT=/usr/local/go
export GOPATH=/tmp/gopath
cd /tmp/gopath/src/github.com/terraform-providers/terraform-provider-cloudfoundry

export CF_USER=admin
export CF_PASSWORD=admin
export CF_SKIP_SSL_VALIDATION=true
export CF_UAA_CLIENT_ID=admin
export CF_API_URL=https://api.$PRIVATE_IP.xip.io
export CF_UAA_CLIENT_SECRET=admin-client-secret

make testacc

EOF

make release

set +e