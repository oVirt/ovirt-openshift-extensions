#!/bin/bash -ex

EXPORTED_ARTIFACTS=exported-artifacts
mkdir -p $EXPORTED_ARTIFACTS

make rpm ARTIFACT_DIR=$EXPORTED_ARTIFACTS

