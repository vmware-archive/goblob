#!/usr/bin/env bash

set -ex

export GOPATH=$PWD/go

ROOT=$(cd $(dirname $0) && pwd)
version=$(cat $ROOT/version/version)

cd go/src/github.com/pivotalservices/goblob
  go build \
    -o $BINARY_FILENAME github.com/pivotalservices/goblob/cmd/goblob \
    -ldflags "-X goblob.Version=${version}"
cd -
