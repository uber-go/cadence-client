#!/bin/bash

set -e

usage () {
    echo 'Installs a specific version of a go-gettable bin to the specified location.'
    echo ''
    echo 'Usage:'
    echo ''
    echo -e "\t$0 [--debug] go-gettable-repo version [path-to-bin-in-repo] install-location"
    echo ''
    echo 'Examples:'
    echo ''
    echo -e "\t$0 go.uber.org/thriftrw 1.10.0 somewhere/thriftrw"
    echo -e '\t\Installs v1.10.0 of go.uber.org/thriftrw to somewhere/thriftrw.'
    echo -e '\t\tNotice that go.uber.org/thriftrw is both the repository AND the binary name.'
    echo ''
    echo -e "\t$0 golang.org/x/lint SOME_SHA golint somewhere/golint"
    echo -e '\t\Installs a specific SHA of e.g. "golang.org/x/lint/golint" to "somewhere/golint",'
    echo -e '\t\tNotice that the golint bin is in a subfolder of the golang.org/x/lint repository.'

    exit 1
}

# needs 3 or 4 args, plus optional --debug

if [ "$1" == "--debug" ]; then
    shift  # consume it
    DEBUG=1
fi

[ $# -ge 3 ] || usage
[ $# -le 4 ] || usage

[ -z $DEBUG ] || set -x

# set up gopath, and make sure it's cleaned up regardless of how we exit
export GOPATH=$(mktemp -d)
trap "rm -rf $GOPATH" EXIT

# variable inits
GO_GETTABLE_REPO="$1"
VERSION="$2"
if [ $# -eq 4 ]; then
    GO_GETTABLE_BIN="$1/$3"
    INSTALL_LOCATION="$4"
elif [ $# -eq 3 ]; then
    GO_GETTABLE_BIN="$1"
    INSTALL_LOCATION="$3"
else
    # should be unreachable
    usage
fi


go get -d "$GO_GETTABLE_BIN"
pushd "$GOPATH/src/$GO_GETTABLE_REPO"
git checkout "$VERSION"

# only glide install when there is a glide file, or it tries to install
# to the current repo (not in our current folder)
if [ -f glide.lock ]; then
    glide install
fi

popd
go build -o "$INSTALL_LOCATION" "$GO_GETTABLE_BIN"
