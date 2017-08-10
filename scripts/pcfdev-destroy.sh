#!/bin/bash

set -e

pushd .test_env
vagrant destroy -f
popd

set +e