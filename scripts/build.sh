#!/bin/bash

set -e

function makeRelease {
    tests/clean.sh
    make testacc
    make release
}

if [[ "$TRAVIS_BRANCH" == "dev" || "$TRAVIS_BRANCH" == "master" && ]] ; then
    echo "Git commit is on dev or master branch - building and running the acceptance tests"
    makeRelease
    exit 0
fi

if [[ "$TRAVIS_PULL_REQUEST" == "true" ]] ; then
    echo "Git commit is a pull request - building and running the acceptance tests"
    makeRelease
    exit 0
fi

if [[ -n "$TRAVIS_TAG" ]] ; then
    echo "Git commit has a release tag - building and running the acceptance tests"
    makeRelease
    exit 0
fi

echo "Git commit is not on dev or master branch or it is not a pull request or it does not have a release tag so acceptance tests will not run."
make build

