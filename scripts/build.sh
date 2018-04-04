#!/bin/bash

if [[ -z "$TRAVIS_TAG" ]] ; then
    echo "Git commit does not have a release tag so acceptance tests will not run."
    
    make build
    exit $?
fi

make testacc
make release
exit $?
