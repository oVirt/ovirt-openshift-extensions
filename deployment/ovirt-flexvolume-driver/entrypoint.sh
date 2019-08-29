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

# Inject credentials into config file
if [ ! -z "$OVIRT_CONNECTION_CREDENTIAL_FILE" ] && [ -f "$OVIRT_CONNECTION_CREDENTIAL_FILE" ] ; then
    grep -q '^username=' "$OVIRT_CONNECTION_CREDENTIAL_FILE" && sed -i '/^username=.*/d' $src/ovirt-flexvolume-driver.conf
    grep -q '^password=' "$OVIRT_CONNECTION_CREDENTIAL_FILE" && sed -i '/^password=.*/d' $src/ovirt-flexvolume-driver.conf
    cat $OVIRT_CONNECTION_CREDENTIAL_FILE >> $src/ovirt-flexvolume-driver.conf
else
    set +x
    if [ ! -z "$OVIRT_CONNECTION_USERNAME" ] ; then
        echo "Applying username from OVIRT_CONNECTION_USERNAME to ovirt-flexvolume-driver.conf"
        grep -q '^username=' $src/ovirt-flexvolume-driver.conf \
            && sed -i "s/^username=.*/username=$OVIRT_CONNECTION_USERNAME/" $src/ovirt-flexvolume-driver.conf \
            || echo "username=$OVIRT_CONNECTION_USERNAME" >> $src/ovirt-flexvolume-driver.conf
    fi
    if [ ! -z "$OVIRT_CONNECTION_PASSWORD" ] ; then
        echo "Applying password from OVIRT_CONNECTION_PASSWORD to ovirt-flexvolume-driver.conf"
        grep -q '^password=' $src/ovirt-flexvolume-driver.conf \
            && sed -i "s/^password=.*/password=$OVIRT_CONNECTION_PASSWORD/" $src/ovirt-flexvolume-driver.conf \
            || echo "password=$OVIRT_CONNECTION_PASSWORD" >> $src/ovirt-flexvolume-driver.conf
    fi
    set -x
fi

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

