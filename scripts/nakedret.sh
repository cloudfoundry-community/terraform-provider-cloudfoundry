#!/usr/bin/env bash

# Check nakedret
echo "==> Checking for naked returns..."

if ! which nakedret > /dev/null; then
    echo "==> Installing nakedret..."
    go get -u github.com/alexkohler/nakedret
fi

function run {
  local path=$1
  local pwd=$(pwd)
  local err=0

  cd $path
  while read line; do
    err=1
    echo "${path}/${line}"
  done < <(nakedret -l 0 2>&1)
  cd $pwd

  return ${err}
}

errs=0
run cloudfoundry
errs=$((errs + $?))
run cloudfoundry/cfapi
errs=$((errs + $?))
run cloudfoundry/repo
errs=$((errs + $?))


if [[ ${errs} -ne 0 ]]; then
    echo ""
    echo "Please remove naked returns. You can check directly with \`make nakedcheck\`"
    exit 1
fi
exit 0
