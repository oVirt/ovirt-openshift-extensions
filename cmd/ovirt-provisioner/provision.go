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
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/ovirt/ovirt-flexdriver/internal"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/kubelet/apis"
)

const (
	flexvolumeVendor = "ovirt"
	flexvolumeDriver = "ovirt-flexdriver"

	// are we allowed to set this? else make up our own
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = "ovirt-provisioner"

	annVolumeID = "ovirt.external-storage.incubator.kubernetes.io/VolumeID"

	annProvisionerID = "Provisioner_Id"
)

// NewOvirtProvisioner creates a new Ovirt provisioner
func NewOvirtProvisioner(ovirtClient *internal.Ovirt) controller.Provisioner {
	var identity types.UID
	provisioner := ovirtProvisioner{
		ovirtClient: ovirtClient,
		identity:    identity,
	}
	return provisioner
}

type ovirtProvisioner struct {
	ovirtClient *internal.Ovirt
	identity    types.UID
}

//func (p *ovirtProvisioner) getAccessModes() []v1.PersistentVolumeAccessMode {
//	return []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
//}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p ovirtProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	glog.Infof("About to provision a disk named %s", options.PVName)
	// call ovirt api, create an unattached disk
	vol, err := p.ovirtClient.CreateUnattachedDisk(
		options.PVName,
		options.Parameters["ovirtStorageDomain"],
		options.PVC.Spec.Size(),
		false, // TODO support the PV Spec access mode?
		options.Parameters["ovirtDiskFormat"],
	)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy
	annotations[annProvisionerID] = string(p.identity)
	annotations[annVolumeID] = vol.Id

	labels := make(map[string]string)
	labels[apis.LabelZoneFailureDomain] = "" // TODO A Zone is ovirt cluster?

	//volSize := volumeOptions.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]

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
				v1.ResourceName(v1.ResourceStorage): *resource.NewQuantity(int64(vol.ProvisionedSize), "gi"),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{

				FlexVolume: &v1.FlexVolumeSource{
					Driver:   fmt.Sprintf("%s/%s", flexvolumeVendor, flexvolumeDriver),
					Options:  map[string]string{},
					ReadOnly: false, // TODO support PV spec access mode?
				},
			},
		},
	}

	return pv, nil
}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p ovirtProvisioner) Delete(volume *v1.PersistentVolume) error {
	glog.Infof("About to delete disk %s id %s", volume.Name, volume.Annotations[annVolumeID])
	// Remove the disk from the storage domain - TODO consider wipe-after-delete and the rest of the options later
	_, err := p.ovirtClient.Delete("disks/" + volume.Annotations[annVolumeID])
	return err
}
