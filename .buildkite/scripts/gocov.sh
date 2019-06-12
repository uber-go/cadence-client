#!/bin/sh

set -ex

# fetch codecov reporting tool
go get github.com/mattn/goveralls

# download cover files from all the tests
mkdir -p .build/coverage
buildkite-agent artifact download .build/coverage/unit_test_cover.out .build/coverage --step unit-test
buildkite-agent artifact download .build/coverage/integ_test_sticky_off_cover.out .build/coverage --step integration-test-sticky-off
buildkite-agent artifact download .build/coverage/integ_test_sticky_on_cover.out .build/coverage --step integration-test-sticky-on

# report coverage
make cover_ci

# cleanup
rm -rf .build