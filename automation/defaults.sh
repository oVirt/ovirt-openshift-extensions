#!/bin/bash -x
#
# General base file for automation actions.

ORG=github.com/ovirt
REPO=ovirt-openshift-extensions
GO_VER=1.9.2

function install_go() {
    curl https://storage.googleapis.com/golang/go${GO_VER}.linux-amd64.tar.gz -o go${GO_VER}.tar.gz
    tar -zxvf go${GO_VER}.tar.gz -C /usr/local/
    export PATH=/usr/local/go/bin:$PATH
}

[[ "$(go version | awk '{print $3}')" != "go${GO_VER}" ]] && install_go