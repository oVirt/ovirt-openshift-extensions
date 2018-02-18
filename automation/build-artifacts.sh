#!/bin/bash -ex

docker build deployment/ovirt-flexdriver

#mkdir -p /tmp/go/{src,pkg,bin}
#export GOPATH=/tmp/go
#export PATH=${PATH}:${GOPATH}/bin

# make sure latest glide is used
#curl https://glide.sh/get | sh

#glide cc
#make deps
#make build-flex
