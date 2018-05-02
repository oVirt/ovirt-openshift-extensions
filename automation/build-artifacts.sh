#!/bin/bash -ex

automation_dir=$(dirname $(readlink -f $0))

set +e
source ${automation_dir}/defaults.sh
set -e

mkdir -p /tmp/build/src/${ORG}
ln -s $HOME /tmp/build/src/${ORG}/
export GOPATH=/tmp/build
cd /tmp/build/src/${ORG}/${REPO}

EXPORTED_ARTIFACTS=exported-artifacts
mkdir -p $EXPORTED_ARTIFACTS

make rpm ARTIFACT_DIR=$EXPORTED_ARTIFACTS

find $EXPORTED_ARTIFACTS -name "*.rpm" | xargs -I '{}' cp -v '{}' .

make container
make container-push
make apb_build apb_docker_push
