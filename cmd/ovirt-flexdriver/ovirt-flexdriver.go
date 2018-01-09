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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ovirt/ovirt-flexdriver/internal"
	"gopkg.in/gcfg.v1"
	"os"
	"strings"
	"time"
)

const usage = `Usage:
	ovirt-flexdriver init
	ovirt-flexdriver attach <json params> <nodename>
	ovirt-flexdriver detach <mount device> <nodename>
	ovirt-flexdriver waitforattach <mount device> <json params>
	ovirt-flexdriver mountdevice <mount dir> <mount device> <json params>
	ovirt-flexdriver unmountdevice <mount dir>
	ovirt-flexdriver isattached <json params> <nodename>
`

var driverConfigFile = "ovirt-flexdriver.conf"

func main() {
	s, e := App(os.Args)
	if e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, s)
}

func App(args []string) (string, error) {

	if len(args) == 0 {
		return "", errors.New(usage)
	}

	var result internal.Response
	var err error

	switch args[0] {
	case "init":
		result, err = initialize()
	case "attach":
		if len(args) < 3 {
			return "", errors.New(usage)
		}
		result, err = attach(args[1], args[2])
	case "waitforattach":
		if len(args) < 3 {
			return "", errors.New(usage)
		}
		result, err = waitForAttach(args[1], args[2])
	case "detach":
		if len(args) < 3 {
			return "", errors.New(usage)
		}
		detach(args[1], args[2])
	default:
		return "", errors.New(usage)
	}

	if err != nil {
		return "", err
	}

	bytes, err := json.Marshal(result)
	return string(bytes), err
}

func initialize() (internal.Response, error) {
	value, exist := os.LookupEnv("OVIRT_FLEXDRIVER_CONF")
	if exist {
		driverConfigFile = value
	}

	_, err := newOvirt()
	if err != nil {
		return internal.FailedResponse, err
	}
	return internal.Response{Status: "success", Capabilities: struct{ Attach string }{"true"}}, nil
}
func newOvirt() (*internal.Ovirt, error) {
	ovirt := internal.Ovirt{}
	err := gcfg.ReadFileInto(&ovirt, driverConfigFile)
	if err != nil {
		return nil, err
	}
	err = ovirt.Authenticate()
	if err != nil {
		return nil, err
	}
	return &ovirt, nil
}

func attach(jsonOpts string, nodeName string) (internal.Response, error) {
	ovirt, err := newOvirt()
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}
	r, e := internal.AttachRequestFrom(jsonOpts)
	if e != nil {
		return internal.FailedResponse, e
	}

	vm, err := ovirt.GetVM(nodeName)
	// 0. validation - Attach size is legal?
	// 1. query if the disk exists
	// 2. if it exist, is it already attached to a VM (perhaps a detach is in progress)
	// 3. if it is attached, is this vm is this node? if not return error.
	// 4. not? create it and attach to the vm

	if err != nil {
		return internal.FailedResponseFromError(err), err
	}
	// vm exist?
	if vm.Id == "" {
		e := errors.New(fmt.Sprintf("VM %s doesn't exist", nodeName))
		return internal.FailedResponseFromError(e), e
	}
	diskResult, err := ovirt.GetDiskByName(r.VolumeName)
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	// 1. no such disk, create it on the VM
	if len(diskResult.Disks) == 0 {
		attachment, err := ovirt.CreateDisk(r.VolumeName, r.StorageDomain, r.Size, r.Mode == "ro", vm.Id)
		if err != nil {
			return internal.FailedResponseFromError(err), err
		}
		return responseFromDiskAttachment(attachment.Id, attachment.Interface), err
	} else {
		// 2. The disk - fetch the disk attachment on the VM
		attachment, err := ovirt.GetDiskAttachment(vm.Id, diskResult.Disks[0].Id)
		if err != nil {
			return internal.FailedResponseFromError(err), err
		}
		return responseFromDiskAttachment(attachment.Id, attachment.Interface), err
	}

	return internal.FailedResponse, err
}

func isattached(jsonOpts string, nodeName string) {
	fmt.Printf("isattached %s %s \n", jsonOpts, nodeName)
}

func detach(mountDevice string, nodeName string) {
	fmt.Printf("detaching %s %s \n", mountDevice, nodeName)
}

// waitForAttach wait for a device disk to be attached to the VM. The disk attachment
// status expected to be true.
// deviceName - the full device name as the output of the #attach call i.e /dev/disk/by-id/virtio-abcdef123
// see 	#responseFromDiskAttachment
// jsonOpts - json string in the form of
func waitForAttach(deviceName string, jsonOpts string) (internal.Response, error) {
	ovirt, err := newOvirt()
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	nodeName := internal.GetOvirtNodeName()
	if nodeName == "" {
		e := errors.New(fmt.Sprintf("Invalid node name '%s'", nodeName))
		return internal.FailedResponseFromError(e), e
	}

	//device name is a path on the os - get the id from it
	id := extractDeviceId(deviceName)

	// FIXME fuzzy get by id since the id is partial
	diskAttachments, err := ovirt.GetDiskAttachments(nodeName)
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	var attachment internal.DiskAttachment
	for _, d := range diskAttachments {
		if strings.HasPrefix(d.Disk.Id, id) {
			attachment = d
		}
	}
	if attachment.Id == "" {
		err = errors.New(fmt.Sprintf("Disk with id '%s' was not found", id))
		return internal.FailedResponseFromError(err), err
	}

	retries := 5
	timeout := time.Second * 10
	for retries > 0 {
		if attachment.Active {
			break
		}
		time.Sleep(timeout)
		attachment, err = ovirt.GetDiskAttachment(nodeName, attachment.Id)
		if err != nil {
			return internal.FailedResponseFromError(err), err
			break
		}
		retries--
	}
	return internal.SuccessfulResponse, nil
}

// extractDeviceId will try to extract the ID of the disk from its path on the OS
// deviceName should be the device path as returned by the attach call.
// Basically revering the responseFromDiskAttachment
func extractDeviceId(deviceName string) string {
	fieldsFunc := strings.FieldsFunc(deviceName, func(r rune) bool { return '/' == r })
	id := fieldsFunc[len(fieldsFunc)-1]
	if strings.HasPrefix(id, "scsi") {
		return strings.TrimPrefix(id, "scsi-0QEMU_QEMU_HARDDISK_")
	}
	if strings.HasPrefix(id, "virtio") {
		return strings.TrimPrefix(id, "virtio-")
	}
	return ""
}

func mountDevice(mountDir string, mountDevice string, jsonOpts string) {
	fmt.Printf("mountDevicee %s \n", mountDevice)
}

func unmountDevice(mountDevice string) {
	fmt.Printf("mountDevicee %s \n", mountDevice)
}

func responseFromDiskAttachment(diskId string, diskInterface string) internal.Response {
	r := internal.SuccessfulResponse
	shortDiskId := diskId[:16]
	switch diskInterface {
	case "virtio":
		r.Device = "/dev/disk/by-id/virtio-" + shortDiskId
	case "virtio_iscsi":
		r.Device = "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_" + shortDiskId
	default:
		return internal.FailedResponseFromError(errors.New("device type is unsupported"))
	}
	return r
}
