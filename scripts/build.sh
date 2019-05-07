#!/bin/bash

set -e

function makeRelease {
    echo "No need for this build ? As a reminder, commits can be instructed to not trigger a travis build by setting the [skip ci keyword](https://docs.travis-ci.com/user/customizing-the-build/#skipping-a-build) in their commit message."
    tests/clean.sh
    make testacc
    make release
}

if [[ -z "$CF_PASSWORD" ]] ; then
    echo "Git commit is probably on a fork - only building it"
    make build
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

if [[ "$TRAVIS_BRANCH" == "dev" || "$TRAVIS_BRANCH" == "master" ]] ; then
    echo "Git commit is on dev or master branch - building and running the acceptance tests"
    makeRelease
    exit 0
fi

echo "Git commit is likely to be preparation work for an upcoming PR so acceptance tests will not run (not on dev or master branch, nor it is not a pull request, nor it does not have a release tag)"
make build

