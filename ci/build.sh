#!/usr/bin/env bash

set -ex

export GOPATH=$PWD/go

version=$(cat $PWD/version/version)

cd go/src/github.com/pivotalservices/goblob
  go build \
    -o $BINARY_FILENAME github.com/pivotalservices/goblob/cmd/goblob \
    -ldflags "-s -w -X goblob.Version=${version}"
cd -
