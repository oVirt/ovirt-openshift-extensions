#!/bin/bash -ex

source defaults.conf

mkdir -p /tmp/build/src/${ORG}
ln -s $HOME /tmp/build/src/${ORG}/
export GOPATH=/tmp/build
cd /tmp/build/src/${ORG}/${REPO}

make build test
