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

# Now that we have it we can just sleep
while true;do sleep 1d;done
