#!/bin/bash -ex

echo building ovirt-provisioner

docker build .
#mkdir -p /tmp/go/{src,pkg,bin}
#export GOPATH=/tmp/go
#export PATH=${PATH}:${GOPATH}/bin

# make sure latest glide is used
#curl https://glide.sh/get | sh

#glide cc
#make deps
#make build-flex
