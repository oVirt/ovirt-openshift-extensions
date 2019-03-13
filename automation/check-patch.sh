#!/bin/bash -ex

EXPORTED_ARTIFACTS=exported-artifacts
mkdir -p $EXPORTED_ARTIFACTS

docker version

make build-containers

# get the build outputs
c=$(docker run --rm -d --entrypoint sleep quay.io/rgolangh/ovirt-flexvolume-driver 5m)
docker cp $c:/tmp/coverage.html $EXPORTED_ARTIFACTS/coverage.html
docker stop $c
