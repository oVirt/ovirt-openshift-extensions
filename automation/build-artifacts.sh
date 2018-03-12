#!/bin/bash -ex

mkdir -p /tmp/build/src/github.com/rgolangh
ln -s $HOME /tmp/build/src/github.com/rgolangh/
export GOPATH=/tmp/build
cd /tmp/build/src/github.com/rgolangh/ovirt-flexdriver

EXPORTED_ARTIFACTS=exported-artifacts
mkdir -p $EXPORTED_ARTIFACTS

make rpm ARTIFACT_DIR=$EXPORTED_ARTIFACTS
find $EXPORTED_ARTIFACTS -name "*.rpm" | xargs -I '{}' cp -v '{}' .

make container