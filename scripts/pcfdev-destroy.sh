#!/bin/bash

pushd .test_env
vagrant destroy -f > /dev/null 2>&1
popd
