#!/bin/bash

pushd .test_env >/dev/null
vagrant destroy -f > /dev/null 2>&1
popd >/dev/null
