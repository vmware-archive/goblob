#!/usr/bin/env bash

set -ex

export GOPATH=$PWD/go

root=$PWD
output_path=$root/$OUTPUT_PATH
version=$(cat $root/version/version)

cd go/src/github.com/pivotal-cf/goblob
  go build \
    -o $output_path \
    -ldflags "-s -w -X goblob.Version=${version}" \
    github.com/pivotal-cf/goblob/cmd/goblob
cd -
