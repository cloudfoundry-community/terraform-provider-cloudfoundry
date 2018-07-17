#!/bin/bash

set -e

GIT_BRANCH=`git symbolic-ref -q --short HEAD 2> /dev/null`

if [[ "$GIT_BRANCH" != "dev" && "$GIT_BRANCH" != "master" && "$TRAVIS_PULL_REQUEST" == "false" && -z "$TRAVIS_TAG" ]] ; then
    echo "Git commit is not on dev or master branch or it is not a pull request or it does not have a release tag so acceptance tests will not run."
    make build
    exit 0
fi

tests/clean.sh
make testacc
make release

