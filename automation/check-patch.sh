#!/bin/bash -ex

mkdir -p /tmp/build/src/github.com/rgolangh
ln -s $HOME /tmp/build/src/github.com/rgolangh/
export GOPATH=/tmp/build
cd /tmp/build/src/github.com/rgolangh/ovirt-flexdriver

make build test