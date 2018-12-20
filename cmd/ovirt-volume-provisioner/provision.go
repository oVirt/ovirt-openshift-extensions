/*
Copyright 2018 oVirt-maintainers

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"strconv"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/kubelet/apis"

	"github.com/ovirt/ovirt-openshift-extensions/internal"
)

const (
	flexvolumeVendor = "ovirt"
	flexvolumeDriver = "ovirt-flexvolume-driver"

	// are we allowed to set this? else make up our own
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = "ovirt-provisioner"
	annVolumeID = "ovirt.external-storage.incubator.kubernetes.io/VolumeID"
	annProvisionerID = "Provisioner_Id"

	parameterStorageDomainName = "ovirtStorageDomain"
	parameterDiskThinProvisioning = "ovirtDiskThinProvisioning"
	parameterFsType = "fsType"
)

// NewOvirtProvisioner creates a new Ovirt provisioner
func NewOvirtProvisioner(ovirtApi internal.OvirtApi) controller.Provisioner {
	var identity types.UID
	provisioner := &ovirtProvisioner{
		ovirtApi: ovirtApi,
		identity: identity,
	}
	return provisioner
}

type ovirtProvisioner struct {
	ovirtApi internal.OvirtApi
	identity types.UID
}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p ovirtProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	// call ovirt api, create an unattached disk
	capacity := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	volSizeBytes := capacity.Value()
	fsType, exists := options.Parameters[parameterFsType]
	if !exists || fsType == "" {
		fsType = "ext4"
	}

	glog.Infof("About to provision a disk name: %s domain: %s size: %v thin provisioned: %s file system: %s",
		options.PVName,
		options.Parameters[parameterStorageDomainName],
		volSizeBytes,
		options.Parameters[parameterDiskThinProvisioning],
		fsType,
	)

	thinProvisioning := true
	val, ok := options.Parameters[parameterDiskThinProvisioning]
	if ok {
		b, e := strconv.ParseBool(val)
		if e != nil {
			glog.Warningf("wrong value for parameter %s. Using default %v", parameterDiskThinProvisioning, thinProvisioning)
		} else {
			thinProvisioning = b
		}
	}

	vol, err := p.ovirtApi.CreateUnattachedDisk(
		options.PVName,
		options.Parameters[parameterStorageDomainName],
		volSizeBytes,
		false, // TODO support the PV Spec access mode?
		thinProvisioning,
	)
	if err != nil {
		return nil, err
	}

	pv := pvFromDisk(p.identity, vol, options, fsType)
	return pv, nil
}

// pvFromDisk takes an ovirt disk details and created a PersistentVolume object
func pvFromDisk(provisionerId types.UID, disk internal.Disk, options controller.VolumeOptions, fsType string) *v1.PersistentVolume {
	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy
	annotations[annProvisionerID] = string(provisionerId)
	annotations[annVolumeID] = disk.Id
	labels := make(map[string]string)
	labels[apis.LabelZoneFailureDomain] = "" // TODO A Zone is ovirt cluster?

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{

			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): *resource.NewQuantity(int64(disk.ProvisionedSize), "gi"),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{

				FlexVolume: &v1.FlexPersistentVolumeSource{
					Driver:   fmt.Sprintf("%s/%s", flexvolumeVendor, flexvolumeDriver),
					Options:  map[string]string{},
					ReadOnly: false, // TODO support PV spec access mode?
					FSType:   fsType,
				},
			},
		},
	}
	return pv
}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p ovirtProvisioner) Delete(volume *v1.PersistentVolume) error {
	glog.Infof("About to delete disk %s id %s", volume.Name, volume.Annotations[annVolumeID])
	// Remove the disk from the storage domain - TODO consider wipe-after-delete and the rest of the options later
	_, err := p.ovirtApi.Delete("disks/" + volume.Annotations[annVolumeID])
	return err
}
