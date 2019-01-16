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

vmId=$(cat /sys/devices/virtual/dmi/id/product_uuid)
if [[ "$vmId" == "" ]]; then
  echo "failed to extract the VM id. Attach actions will fail. Exiting"
  exit 1
fi

# append per node values to the config
echo "ovirtVmId=${vmId}" >> $src/ovirt-flexvolume-driver.conf

# remove the old config directory
rm -v -rf $dest/ovirt~ovirt-flexvolume-driver

# to prevent half-baked initialization do an atomic move into $dest
# the 'ovirt~' prefix is added per flex specification
mv -v $src $dest/ovirt~ovirt-flexvolume-driver

# Now that we have it we can just sleep
while true;do sleep 1d;done

