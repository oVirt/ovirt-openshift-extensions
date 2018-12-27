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

import "errors"

// ovirt api common errors
var (
	ErrNotExist = errors.New("resource does not exists")
)

type DiskAttachment struct {
	Id          string `json:"id,omitempty"`
	Bootable    bool   `json:"bootable,string"`
	PassDiscard bool   `json:"pass_discard,string"`
	Interface   string `json:"interface,omitempty"`
	Active      bool   `json:"active,string"`
	Disk        Disk   `json:"disk"`
	ReadOnly    bool   `json:"read_only,string"`
}

type DiskAttachmentResult struct {
	DiskAttachments []DiskAttachment `json:"disk_attachment"`
}

type DiskFormat string
type Sparse		bool

type Disk struct {
	Id              string         `json:"id,omitempty"`
	Name            string         `json:"name"`
	ActualSize      uint64         `json:"actual_size,omitempty,string"`
	ProvisionedSize uint64         `json:"provisioned_size,string"`
	Status          string         `json:"status,omitempty"`
	Format          DiskFormat     `json:"format"`
	StorageDomains  StorageDomains `json:"storage_domains"`
	Sparse  		Sparse         `json:"sparse,string"`

}

type DiskResult struct {
	Disks []Disk `json:"disk"`
}

type StorageDomains struct {
	Domains []StorageDomain `json:"storage_domain"`
}

type StorageDomain struct {
	Name string `json:"name"`
	Storage struct{  Type string `json:"type,omitempty"`} `json:"storage,omitempty,"`
}

type VM struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Fqdn string `json:"fqdn"`
	Nics struct { Nics []Nic `json:"nic"` } `json:"nics"`
	Status string `json:"status"`
}

type Nic struct {
	Interface string `json:"interface"`
	Linked bool `json:"linked:string"`
	Devices struct { Devices []Device `json:"reported_device"` } `json:"reported_devices"`
}

type Device struct {
	Ips struct { Ips []Ip `json:"ip"` } `json:"ips"`
}

type Ip struct {
	Address string `json:"address"`
	// v4, v6
	Version string `json:"version"`
}

type VMResult struct {
	Vms []VM `json:"vm"`
}

