#!/bin/bash

set -e
set -x

echo $TRAVIS_BRANCH

if [[ "$TRAVIS_BRANCH" != "dev" && "$TRAVIS_BRANCH" != "master" && "$TRAVIS_PULL_REQUEST" == "false" && -z "$TRAVIS_TAG" ]] ; then
    echo "Git commit is not on dev or master branch or it is not a pull request or it does not have a release tag so acceptance tests will not run."
    make build
    exit 0
fi

tests/clean.sh
make testacc
make release

