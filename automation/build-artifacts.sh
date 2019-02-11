#!/bin/bash -ex

EXPORTED_ARTIFACTS=exported-artifacts
mkdir -p $EXPORTED_ARTIFACTS

make build-containers ARTIFACT_DIR=$EXPORTED_ARTIFACTS
make push-containers
make apb_build apb_docker_push

cp containers-artifacts.list $EXPORTED_ARTIFACTS || true
# get the build outputs
c=$(docker run --rm -d --entrypoint sleep quay.io/rgolangh/ovirt-flexvolume-driver 5m)
docker cp $c:/tmp/coverage.html $EXPORTED_ARTIFACTS/coverage.html
docker stop $c
