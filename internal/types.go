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

type DiskAttachment struct {
	Id          string `json:"id,omitempty"`
	Bootable    bool   `json:"bootable,string"`
	PassDiscard bool   `json:"pass_discard,string"`
	Interface   string `json:"interface"`
	Active      bool   `json:"active,string"`
	Disk        Disk   `json:"disk"`
	ReadOnly    bool   `json:"read_only,string"`
}

type DiskAttachmentResult struct {
	DiskAttachments []DiskAttachment `json:"disk_attachment"`
}

type DiskFormat string

const RAW DiskFormat = "raw"
const COW DiskFormat = "cow"

type Disk struct {
	Id              string         `json:"id,omitempty"`
	Name            string         `json:"name"`
	ActualSize      uint64         `json:"actual_size"`
	ProvisionedSize uint64         `json:"provisioned_size"`
	Status          string         `json:"status,omitempty"`
	Format          DiskFormat     `json:"format"`
	StorageDomains  StorageDomains `json:"storage_domains"`
}

type DiskResult struct {
	Disks []Disk `json:"disk"`
}

type StorageDomains struct {
	Domains []StorageDomain `json:"storage_domain"`
}

type StorageDomain struct {
	Name string `json:"name"`
}

type VM struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type VMResult struct {
	Vms []VM `json:"vm"`
}
