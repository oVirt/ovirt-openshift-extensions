#!/bin/bash -ex

EXPORTED_ARTIFACTS=exported-artifacts
mkdir -p $EXPORTED_ARTIFACTS

make container ARTIFACT_DIR=$EXPORTED_ARTIFACTS
make container-push
make apb_build apb_docker_push
