/*
Copyright 2017 oVirt-maintainers

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

package internal

import (
	"encoding/json"
	"strings"
)

type Status string

const (
	Success      Status = "Success"
	Failure      Status = "Failure"
	NotSupported Status = "Not supported"
)

var (
	FailedResponse       = Response{Status: Failure}
	SuccessfulResponse   = Response{Status: Success, Message: "success"}
	NotSupportedResponse = Response{Status: NotSupported}
)

type OvirtApi interface {
	Authenticate() error
	Get(path string) ([]byte, error)
	Post(path string, data interface{}) (string, error)
	Delete(path string) ([]byte, error)
	GetVM(name string) (VM, error)
	GetVMById(id string) (VM, error)
	GetVMs(query string) ([]VM, error)
	GetDiskAttachment(vmId, diskId string) (DiskAttachment, error)
	GetDiskAttachments(vmId string) ([]DiskAttachment, error)
	DetachDiskFromVM(vmId string, diskId string) error
	GetDiskByName(diskName string) (DiskResult, error)
	CreateUnattachedDisk(diskName string, storageDomainName string, sizeIbBytes int64, readOnly bool, thinProvisioning bool) (Disk, error)
	CreateDisk(
		diskName string,
		storageDomainName string,
		readOnly bool,
		vmId string,
		diskId string,
		diskInterface string) (DiskAttachment, error)
	GetConnectionDetails() Connection
}

type Response struct {
	Status       Status        `json:"status"`               //"status": "<Success/Failure/Not supported>",
	Message      string        `json:"message"`              //"message": "<Reason for success/failure>",
	Device       string        `json:"device,omitempty"`     //"device": "<Path to the device attached. This field is valid only for attach & waitforattach call-outs>"
	VolumeName   string        `json:"volumeName,omitempty"` //"volumeName": "<Cluster wide unique name of the volume. Valid only for getvolumename call-out>"
	Attached     bool          `json:"attached,omitempty"`   //"attached": <True/False (Return true if volume is attached on the node. Valid only for isattached call-out)>
	Capabilities *Capabilities `json:",omitempty"`
}

type Capabilities struct {
	//"capabilities": <Only included as part of the Init response>
	Attach bool `json:"attach,omitempty"` //: <True/False (Return true if the driver implements attach and detach)>
}

type AttchResponse struct {
	Response
}

type AttachRequest struct {
	StorageDomain string `json:"ovirtStorageDomain"`
	VolumeName    string `json:"kubernetes.io/pvOrVolumeName,omitempty"`
	Size          string `json:"capacity,omitempty"`
	FsType        string `json:"kubernetes.io/fsType"`
	Mode          string `json:"kubernetes.io/readwrite"`
	// TODO use k8s secret?
	Secret     string `json:"kubernetes.io/secret,omitempty"`
	VolumeId   string `json:"volumeID,omitempty"`
	CustomSize string `json:"size,omitempty"`
}

func AttachRequestFrom(s string) (AttachRequest, error) {
	r := AttachRequest{}
	err := json.Unmarshal([]byte(s), &r)
	return r, err
}

func FailedResponseFromError(e error, more ...string) Response {
	r := Response{Status: Failure}
	r.Message = e.Error() + strings.Join(more, " ")
	return r
}
