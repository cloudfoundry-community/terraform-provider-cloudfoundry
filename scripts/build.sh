#!/bin/bash

if [[ "$TRAVIS_PULL_REQUEST" == "false" && -z "$TRAVIS_TAG" ]] ; then
    echo "Git commit is not a pull request or it does not have a release tag so acceptance tests will not run."
    
    make build
    exit $?
fi

make testacc
make release
exit $?
