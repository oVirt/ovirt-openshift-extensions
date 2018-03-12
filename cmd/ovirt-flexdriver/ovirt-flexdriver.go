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
	"github.com/rgolangh/ovirt-flexdriver/internal"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"gopkg.in/gcfg.v1"
	"os"
	"os/exec"
	"path/filepath"
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
	ovirt-flexdriver getvolumename <json params>
`

var driverConfigFile string
var ovirtVmName string

func main() {
	s, e := App(os.Args[1:])
	log, err := logging.NewSyslogBackend("")
	if err == nil && log != nil {
		// syslog may not be available, i.e in busybox
		log.Writer.Info("invoking driver ----- " + strings.Join(os.Args, ","))
		defer log.Writer.Info("invoking driver ----- " + s)
	}
	if e != nil {
		fmt.Fprint(os.Stderr, e.Error())
		os.Exit(1)
	}
	fmt.Fprint(os.Stdout, s)
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
		result, err = Attach(args[1], args[2])
	case "waitforattach":
		if len(args) < 3 {
			return "", errors.New(usage)
		}
		result, err = WaitForAttach(args[1], args[2])
	case "isattached":
		if len(args) < 3 {
			return "", errors.New(usage)
		}
		result, err = IsAttached(args[1], args[2])
	case "detach":
		if len(args) < 3 {
			return "", errors.New(usage)
		}
		result, err = Detach(args[1], args[2])
	case "mountdevice":
		if len(args) < 4 {
			return "", errors.New(usage)
		}
		result, err = MountDevice(args[1], args[2], args[3])
	case "provision":
		if len(args) < 3 {
			return "", errors.New(usage)
		}
		result, err = Provision(args[1])
	case "delete":
		if len(args) < 3 {
			return "", errors.New(usage)
		}
		result, err = Delete(args[1])
	case "unmountdevice", "unmount":
		if len(args) != 2 {
			return "", errors.New(usage)
		}
		result, err = UnmountDevice(args[1])
	case "getvolumename", "mount":
		result, err = internal.NotSupportedResponse, nil
	default:
		return "", errors.New(usage)
	}

	bytes, marshalingErr := json.Marshal(result)
	if marshalingErr != nil {
		return "", marshalingErr
	}
	return string(bytes), err
}

func initialize() (internal.Response, error) {
	_, err := newOvirt()
	if err != nil {
		return internal.FailedResponse, err
	}
	r := internal.SuccessfulResponse
	r.Capabilities = &internal.Capabilities{Attach: true}
	return r, nil
}

func newOvirt() (*internal.Ovirt, error) {
	value, exist := os.LookupEnv("OVIRT_FLEXDRIVER_CONF")
	if exist {
		driverConfigFile = value
	} else {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		driverConfigFile = dir + "/ovirt-flexdriver.conf"
	}
	driver := struct {
		internal.Ovirt
		General struct {
			OvirtVmName string `gcfg:"ovirtVmName"`
		}
	}{}
	err := gcfg.ReadFileInto(&driver, driverConfigFile)
	if err != nil {
		err = errors.New(err.Error() + " file is " + driverConfigFile)
		return nil, err
	}
	err = driver.Authenticate()
	if err != nil {
		return nil, err
	}
	ovirtVmName = driver.General.OvirtVmName
	return &driver.Ovirt, nil
}

// Attach will attach the volume to the nodeName.
// If the volume(ovirt's disk) doesn't exist, create it.
// If it exist, try to attach it to the VM
// jsonOpts - contains the volume spec, like name, size etc
// nodeName - k8s nodeName, needs conversion into ovirt's VM
func Attach(jsonOpts string, nodeName string) (internal.Response, error) {
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

	diskResult, err := ovirt.GetDiskByName(fromk8sNameToOvirt(r.VolumeName))
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	if len(diskResult.Disks) == 0 {
		return internal.FailedResponseFromError(fmt.Errorf("disk %s doesn't exist", r.VolumeName)), err
	}

	// fetch the disk attachment on the VM
	attachment, err := ovirt.GetDiskAttachment(vm.Id, diskResult.Disks[0].Id)
	if err != nil {
		_, noAttachment := err.(internal.NotFound)
		if noAttachment {
			attachment, err =
				ovirt.CreateDisk(fromk8sNameToOvirt(r.VolumeName), r.StorageDomain, r.Mode == "ro", vm.Id, diskResult.Disks[0].Id, "virtio_scsi")
			if err != nil {
				return internal.FailedResponseFromError(err), err
			}
		} else {
			return internal.FailedResponseFromError(err), err
		}
	}
	return responseFromDiskAttachment(attachment.Id, attachment.Interface), err
}

// IsAttached will check if the disk exists on the VM attachments collections.
// it will also reply with false in case the vm or the disk do not exist.
func IsAttached(jsonOpts string, nodeName string) (internal.Response, error) {
	ovirt, err := newOvirt()
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}
	r, e := internal.AttachRequestFrom(jsonOpts)
	if e != nil {
		return internal.FailedResponse, e
	}

	vm, err := ovirt.GetVM(nodeName)
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}
	// vm exist?
	if vm.Id == "" {
		e := errors.New(fmt.Sprintf("VM %s doesn't exist", nodeName))
		return internal.FailedResponseFromError(e), e
	}

	// disk exists?
	diskResult, err := ovirt.GetDiskByName(fromk8sNameToOvirt(r.VolumeName))
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	// fetch attachment
	attachment, err := ovirt.GetDiskAttachment(vm.Id, diskResult.Disks[0].Id)
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	result := internal.SuccessfulResponse
	result.Attached = attachment.Id != ""
	return result, nil
}

// Detach will detach the disk from the VM.
// volumeName is a cluster wide unique name of the volume and needs to be converted to ovirt's disk name/id
// nodeName - the hostname with the volume attached. Needs to be converted to ovirt's VM. See #internal.GetOvirtNodeName
func Detach(volumeName string, nodeName string) (internal.Response, error) {
	if nodeName == "" {
		e := errors.New(fmt.Sprintf("Invalid node name '%s'", nodeName))
		return internal.FailedResponseFromError(e), e
	}
	if volumeName == "" {
		e := errors.New(fmt.Sprintf("Invalid volume name '%s'", volumeName))
		return internal.FailedResponseFromError(e), e
	}

	ovirt, err := newOvirt()
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	ovirtDiskName := fromk8sNameToOvirt(volumeName)

	vm, err := ovirt.GetVM(nodeName)
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	diskResult, err := ovirt.GetDiskByName(ovirtDiskName)
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	if len(diskResult.Disks) == 0 {
		//TODO is this an error or ok state for detach?
		err = errors.New(fmt.Sprintf("Disk by name %s does not exist", ovirtDiskName))
		return internal.FailedResponseFromError(err), err
	} else {
		err := ovirt.DetachDiskFromVM(vm.Id, diskResult.Disks[0].Id)
		if err != nil {
			return internal.FailedResponseFromError(err), err
		}
		return internal.SuccessfulResponse, nil
	}
}

// WaitForAttach wait for a device disk to be attached to the VM. The disk attachment
// status expected to be true.
// deviceName - the full device name as the output of the #attach call i.e /dev/disk/by-id/virtio-abcdef123
// see 	#responseFromDiskAttachment
// jsonOpts - json string in the form of
func WaitForAttach(deviceName string, jsonOpts string) (internal.Response, error) {
	ovirt, err := newOvirt()
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	//device name is a path on the os - get the id from it
	id := extractDeviceId(deviceName)

	vm, e := ovirt.GetVM(ovirtVmName)
	if e != nil {
		return internal.FailedResponseFromError(e), e
	}
	// FIXME fuzzy get by id since the id is partial
	diskAttachments, err := ovirt.GetDiskAttachments(vm.Id)
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
	for retries > 0 && !attachment.Active {
		time.Sleep(timeout)
		attachment, err = ovirt.GetDiskAttachment(ovirtVmName, attachment.Id)
		if err != nil {
			return internal.FailedResponseFromError(err), err
		}
		retries--
	}

	if !attachment.Active {
		err = errors.New("disk is not active yet")
		return internal.FailedResponseFromError(err), err
	}
	r := internal.SuccessfulResponse
	r.Device, err = getDeviceNameFromSerial(deviceName)
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}
	return internal.SuccessfulResponse, nil
}

// getDeviceNameFromSerial takes the serial of the disk i.e /dev/disk/by-id/... and returns the link to the /dev
func getDeviceNameFromSerial(deviceSerialIdentifier string) (string, error) {
	deviceRealPath, err := filepath.EvalSymlinks(deviceSerialIdentifier)
	if err != nil {
		return "", err
	}
	return deviceRealPath, nil
}

// extractDeviceId will try to extract the ID of the disk from its path on the OS
// deviceName should be the device path as returned by the attach call.
// Basically revering the responseFromDiskAttachment
func extractDeviceId(deviceName string) string {
	if deviceName == "" {
		return ""
	}
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

// MountDevice mount the mountDevice onto mountDir
// mountDir - should exist already and is not the responsibility of the driver to create
// _ 		- Currently the 2nd argument is unknown and undocumented
// jsonOpts - the regular driver options that containes the name of the volume.
// node - the implementation should use the volume name and to trace the disk device from it in order to mount
func MountDevice(mountDir string, _ string, jsonOpts string) (internal.Response, error) {
	r, err := internal.AttachRequestFrom(jsonOpts)
	if err != nil {
		return internal.FailedResponse, err
	}

	// get the underlying device from the volume name (which is the ovirt disk name)
	response, err := GetVolumeName(jsonOpts)
	if err != nil {
		return response, err
	}

	// Get the real disk device name from device
	device, err := getDeviceNameFromSerial(response.Device)
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	// is there a filesystem on this device?
	filesystem, e := getDeviceInfo(device)
	if e != nil {
		return internal.FailedResponseFromError(e, device), e
	}
	if filesystem == "" {
		// no filesystem - create it
		makeFSErr := makeFS(device, r.FsType)
		if makeFSErr != nil {
			return internal.FailedResponseFromError(makeFSErr), makeFSErr
		}
	}

	cmd := exec.Command("mount", "-t", r.FsType, device, mountDir)
	out, err := cmd.Output()
	exitError, incompleteCmd := err.(*exec.ExitError)
	if err != nil && incompleteCmd {
		return internal.FailedResponseFromError(err, " mount failed with "+string(exitError.Error())), err
	}
	retVal := internal.SuccessfulResponse
	retVal.Message = string(out)
	return retVal, nil
}

func makeFS(device string, fsType string) error {
	// caution, use -F to force creating the filesystem if it doesn't exit. May not be portable for fs other
	// than ext family
	var force string
	if strings.HasPrefix(fsType, "ext") {
		force = "-F"
	}
	cmd := exec.Command("mkfs", force, "-t", fsType, device)
	err := cmd.Run()
	exitError, incompleteCmd := err.(*exec.ExitError)
	if err != nil && incompleteCmd {
		return errors.New(err.Error() + " mkfs failed with " + string(exitError.Error()))
	}
	return nil
}

// UnmountDevice umounts the directory from this node, if its a real mount-point. Otherwise ignore it.
func UnmountDevice(mountDir string) (internal.Response, error) {
	if !isMountpoint(mountDir) {
		// nothing to do, return.
		return internal.SuccessfulResponse, nil
	}

	cmd := exec.Command("umount", mountDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}
	return internal.SuccessfulResponse, nil
}

// isMountpoint find out if a given directory is a real mount point
func isMountpoint(mountDir string) bool {
	cmd := exec.Command("findmnt", mountDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func responseFromDiskAttachment(diskId string, diskInterface string) internal.Response {
	r := internal.SuccessfulResponse
	shortDiskId := diskId[:20]
	switch diskInterface {
	case "virtio":
		r.Device = "/dev/disk/by-id/virtio-" + shortDiskId
	case "virtio_scsi":
		r.Device = "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_" + shortDiskId
	default:
		return internal.FailedResponseFromError(errors.New("device type is unsupported"))
	}
	return r
}

func GetVolumeName(jsonOpts string) (internal.Response, error) {
	ovirt, err := newOvirt()
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}
	jsonArgs, e := internal.AttachRequestFrom(jsonOpts)
	if e != nil {
		return internal.FailedResponse, e
	}

	vm, err := ovirt.GetVM(ovirtVmName)
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}
	// vm exist?
	if vm.Id == "" {
		e := errors.New(fmt.Sprintf("VM %s doesn't exist", ovirtVmName))
		return internal.FailedResponseFromError(e), e
	}
	diskResult, err := ovirt.GetDiskByName(fromk8sNameToOvirt(jsonArgs.VolumeName))
	if err != nil {
		return internal.FailedResponseFromError(err), err
	}

	if len(diskResult.Disks) == 0 {
		//noDisk := errors.New(fmt.Sprintf("Volume with name %s doesn't exist in ovirt", jsonArgs.VolumeName))
		//return internal.FailedResponseFromError(noDisk), noDisk
		// maybe just return the name of the disk as is to indicate it is free?
		return internal.NotSupportedResponse, nil
	} else {
		// fetch the disk attachment on the VM
		attachment, err := ovirt.GetDiskAttachment(vm.Id, diskResult.Disks[0].Id)
		if err != nil {
			err = errors.New(fmt.Sprintf("The volume %s is not attched to the node %s", jsonArgs.VolumeName, ovirtVmName))
			return internal.FailedResponseFromError(err), err
		}
		return responseFromDiskAttachment(attachment.Id, attachment.Interface), err
	}
}

func Delete(s string) (internal.Response, error) {
	return internal.SuccessfulResponse, nil
}
func Provision(s string) (internal.Response, error) {
	return internal.SuccessfulResponse, nil
}

// fromk8sNameToOvirt takes name with '~' and replaces it with '_'
func fromk8sNameToOvirt(s string) string {
	return strings.Replace(s, "~", "_", -1)
}

// getDeviceInfo will return the first Device which is a partition and its filesystem.
// if the given Device disk has no partition then an empty zero valued device will return
func getDeviceInfo(device string) (string, error) {
	cmd := exec.Command("lsblk", "-nro", "FSTYPE", device)
	out, err := cmd.Output()
	exitError, incompleteCmd := err.(*exec.ExitError)
	if err != nil && incompleteCmd {
		return "", errors.New(err.Error() + "lsblk failed with " + string(exitError.Stderr))
	}

	reader := bufio.NewReader(bytes.NewReader(out))
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	return string(line), nil
}
