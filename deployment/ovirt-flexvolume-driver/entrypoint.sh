#!/bin/sh -ex

#
# NOTE: The products this containers installats are expected to be in /opt/ovirt-flexvolume-driver/ folder
#
# The flexvolume driver entrypoint makes sure that the binary exists in the $dest
# and is suited to the deploying the driver as a pod only to expose a hostPath. It is
# form of installing the binary from this container, onto the host that is running it.
#
# Install binaries and conf into the exposed hostPath 

src=$(mktemp -d)
cp -r /opt/ovirt-flexvolume-driver/* $src
dest=/var/tmp/usr/libexec/kubernetes/kubelet-plugins/volume/exec/

cp /usr/bin/ovirt-flexvolume-driver $src/
# append per node values to the config
echo "ovirtVmName=$OVIRT_VM_NAME" >> $src/ovirt-flexvolume-driver.conf

# to prevent half-baked initialization do an atomic move into $dest
# the 'ovirt~' prefix is added per flex specification
mv $src $dest/ovirt~ovirt-flexvolume-driver

# Now that we have it we can just sleep
while true;do sleep 1d;done
