#!/bin/sh -ex

dir=/usr/libexec/kubernetes/kubelet-plugins/volume/exec/

# The flexvolume driver entrypoint makes sure that the binary exists in the $dir
# and is suited to the deploying the driver as a pod only to expose a hostPath. It is
# form of installing the binary from this container, onto the host that is running it.
#
# The rpm is probably installed during container build time, but when the host is binding $dir the os
# doesn't see it anymore, but only what's on the host's dir.
# We then perform install again to make sure the binary is there and visible:
ls -la $dir # debug info
rpm -ivh /root/ovirt-flexvolume-driver*.rpm --force

# copy the conf file into the target dir
cp -v /opt/ovirt-flexvolume-driver/ovirt-flexvolume-driver.conf $dir/ovirt~ovirt-flexvolume-driver/

# append per node values to the config
echo "ovirtVmName=$OVIRT_VM_NAME" >> $dir/ovirt~ovirt-flexvolume-driver/ovirt-flexvolume-driver.conf

# to prevent half-baked initilization do and atomic move into $dir
tmpdir=$(mktemp -d)
mv $dir/ovirt~ovirt-flexvolume-driver $tmpdir/
sleep 5
mv $tmpdir/ovirt~ovirt-flexvolume-driver $dir
rmdir $tmpdir

# Now that we have it we can just sleep
while true;do sleep 1d;done
