#!/bin/bash -ex

automation_dir=$(dirname $(readlink -f $0))
source ${automation_dir}/defaults.sh

mkdir -p /tmp/build/src/${ORG}
ln -s $HOME /tmp/build/src/${ORG}/
export GOPATH=/tmp/build
cd /tmp/build/src/${ORG}/${REPO}

make build test
