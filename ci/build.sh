#!/usr/bin/env bash

set -ex

export GOPATH=$PWD/go

root=$PWD
output_path=$root/$OUTPUT_PATH
version=$(cat $root/version/version)

cd go/src/github.com/pivotalservices/goblob
  go build \
    -o $output_path github.com/pivotalservices/goblob/cmd/goblob \
    -ldflags "-s -w -X goblob.Version=${version}"
cd -
