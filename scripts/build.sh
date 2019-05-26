#!/bin/bash

set -e

function testAcceptance {
    tests/clean.sh
    if [[ "$TRAVIS_BUILD_STAGE_NAME" == "Testds" ]] ; then
        TESTARGS="-run 'TestAccData'" make testacc
    fi
    if [[ "$TRAVIS_BUILD_STAGE_NAME" == "Testimport" ]] ; then
        TESTARGS="-run 'TestAcc.*_import.*'" make testacc
    fi
    if [[ "$TRAVIS_BUILD_STAGE_NAME" == "Testres" ]] ; then
        TESTARGS="-run 'TestAccRes'" make testacc
    fi
    if [[ "$TRAVIS_BUILD_STAGE_NAME" == "Testlint" ]] ; then
        make check
    fi
}

if [ -z "$CF_PASSWORD" ]; then
    echo "Git commit is probably on a fork - only building it"
    make build
    exit 0
fi

if [[ "$TRAVIS_BUILD_STAGE_NAME" == "Deploy" ]] ; then
    make release
    exit 0
fi

if [[ "$TRAVIS_PULL_REQUEST" == "true" ]] ; then
    echo "Git commit is a pull request - building and running the acceptance tests"
    testAcceptance
    exit 0
fi

if [[ -n "$TRAVIS_TAG" ]] ; then
    echo "Git commit has a release tag - building and running the acceptance tests"
    testAcceptance
    exit 0
fi

if [[ "$TRAVIS_BRANCH" == "dev" || "$TRAVIS_BRANCH" == "master" ]] ; then
    echo "Git commit is on dev or master branch - building and running the acceptance tests"
    testAcceptance
    exit 0
fi

echo "Git commit is likely to be preparation work for an upcoming PR so acceptance tests will not run (not on dev or master branch, nor it is not a pull request, nor it does not have a release tag)"
make build

