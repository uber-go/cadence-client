#!/bin/bash

set -e

# In a nutshell, this script:
# - makes a tempdir and moves to it
# - go gets the requested bin (but does not install it)
# - cds to the repo
# - checks out the requested version
# - maybe runs `glide install`
# - builds the bin and puts it where you told it to.
#
# Since doing that verbatim is a bit noisy, and pinning tools tends to
# cause them to slowly get out of date, it does two additional things:
# - suppresses output unless `--debug` is passed
# - checks for newer commits / tags than the one you passed
# - print the newer sha/version (if any) to stderr so it's always visible

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
else
    # otherwise, redirect to /dev/null (requires eval)
    TO_DEV_NULL='>/dev/null'
    # pass quiet flags where needed (do not quote)
    QUIET='--quiet'
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
# eval so redirection works when quiet
eval "pushd $GOPATH/src/$GO_GETTABLE_REPO $TO_DEV_NULL"

HEAD=$(git rev-parse HEAD)

# silently check out, reduces a lot of default spam
git checkout $QUIET "$VERSION"

if [ -z "$(echo "$VERSION" | grep -E '^v{0,1}\d+(\.\d+){0,3}$')" ]; then
    # not versioned, check for newer commits
    BEHIND=$(git rev-list "..$HEAD" | wc -l | tr -dc '0-9')
    [ "$BEHIND" -eq 0 ] || (>&2 echo "$GO_GETTABLE_REPO is $BEHIND commits behind the current HEAD: $HEAD")
else
    # versioned, check for newer tags.
    # using xargs because it's safer (for big lists) + it doesn't result in
    # a ton of SHAs in debug output like `git describe --tags $(...)` causes.
    #
    # in brief:
    # - get all tagged shas
    # - get their tags
    # - grep for `v1.2.3.4` since there are a ton of non-release-y ones out there
    # - sort by version, and grab the "biggest"
    LATEST=$(
        git rev-list --tags \
        | xargs git describe --tags 2>/dev/null \
        | grep -E '^v{0,1}\d+(\.\d+){0,3}$' \
        | sort -Vr \
        | head -n 1
    )
    # use sort to check if VERSION >= LATEST
    echo -e "$VERSION\n$LATEST" | sort -Vrc 2>/dev/null || (>&2 echo "$GO_GETTABLE_REPO has a newer tag: $LATEST")
fi

# only glide install when there is a glide file, or it tries to install
# to the current repo (not in our current folder)
if [ -f glide.lock ]; then
    glide $QUIET install
fi

# eval so redirection works when quiet
eval "popd $TO_DEV_NULL"

go build -o "$INSTALL_LOCATION" "$GO_GETTABLE_BIN"
