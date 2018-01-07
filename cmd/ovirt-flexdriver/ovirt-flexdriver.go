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
	// get vm id by name
	if len(diskResult.Disks) == 0 {
		attachment, err := ovirt.CreateDisk(r.VolumeName, r.StorageDomain, r.Size, r.Mode == "ro", vm.Id)
		if err != nil {
			return internal.FailedResponseFromError(err), err
		}
		return responseFromDiskAttachment(attachment), err
	} else {
		attached := internal.SuccessfulResponse
		attached.Device = diskResult.Disks[0].Id
		return attached, nil
	}

	return internal.FailedResponse, err
}

func isattached(jsonOpts string, nodeName string) {
	fmt.Printf("isattached %s %s \n", jsonOpts, nodeName)
}

func detach(mountDevice string, nodeName string) {
	fmt.Printf("detaching %s %s \n", mountDevice, nodeName)
}

func waitForAttach(mountDevice string, nodeName string) {
	fmt.Printf("waitForAttach %s %s \n", mountDevice, nodeName)
}

func mountDevice(mountDir string, mountDevice string, jsonOpts string) {
	fmt.Printf("mountDevicee %s \n", mountDevice)
}

func unmountDevice(mountDevice string) {
	fmt.Printf("mountDevicee %s \n", mountDevice)
}

func responseFromDiskAttachment(d internal.DiskAttachment) internal.Response {
	r := internal.SuccessfulResponse
	id, _ := deviceIdFromVmDiskId(d)
	r.Device = id
	return r
}

func deviceIdFromVmDiskId(attachment internal.DiskAttachment) (string, error) {
	shortDiskId := attachment.Id[:16]
	switch attachment.Interface {
	case "virtio":
		return "/dev/disk/by-id/virtio-" + shortDiskId, nil
	case "virtio_iscsi":
		return "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_" + shortDiskId, nil
	default:
		return "", errors.New("device type is unsupported")
	}
}
