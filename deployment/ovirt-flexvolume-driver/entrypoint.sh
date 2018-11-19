#!/bin/sh -ex

# Expected files location:
#   - binaries under /usr/bin
#   - conf loaded into the container under /opt/ovirt-flexvolume-driver
# 
# This entrypoint makes sure that the binary is copied to the $dest
# and is suited to the deploying the driver as a pod only to expose a hostPath. It is
# form of installing the binary from this container, onto the host that is running it.

src=$(mktemp -d)
dest=/usr/libexec/kubernetes/kubelet-plugins/volume/exec/

cp -v /opt/ovirt-flexvolume-driver/ovirt-flexvolume-driver.conf $src/
cp -v /usr/bin/ovirt-flexvolume-driver $src/

# append per node values to the config
echo "ovirtVmName=$OVIRT_VM_NAME" >> $src/ovirt-flexvolume-driver.conf

# to prevent half-baked initialization do an atomic move into $dest
# the 'ovirt~' prefix is added per flex specification
mv $src $dest/ovirt~ovirt-flexvolume-driver

# Now that we have it we can just sleep
while true;do sleep 1d;done
