package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestExtractDeviceIdForVIRTIO(t *testing.T) {
	ut := "/dev/disk/by-id/virtio-8deb6495-9121-44b9-a"
	expected := "8deb6495-9121-44b9-a"
	id := extractDeviceId(ut)
	if expected != id {
		t.Errorf("expected %s got %s", expected, id)
	}
}

func TestExtractDeviceIdForVIRTIOSCSI(t *testing.T) {
	ut := "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_6a52e54c-003b-45d6-b"
	expected := "6a52e54c-003b-45d6-b"
	id := extractDeviceId(ut)
	if expected != id {
		t.Errorf("expected %s got %s", expected, id)
	}
}

func TestExtractDeviceIdForUnknown(t *testing.T) {
	ut := "/dev/disk/by-id/unknown-type"
	expected := ""
	id := extractDeviceId(ut)
	if expected != id {
		t.Errorf("expected %s got %s", expected, id)
	}
}

func TestGetDeviceFileSystem(t *testing.T) {
	input, err := getDeviceForTest()
	if err != nil {
		t.Fatal(err.Error())
	}
	filesystem, e := getDeviceInfo(input)
	t.Logf("fs %s error %v", filesystem, e)
	if filesystem == "" {
		t.Errorf("expecting some filesystem but got %s", filesystem)
	}
}
func getDeviceForTest() (string, error) {
	cmd := exec.Command("df", "-P", ".")
	output, e := cmd.Output()
	if e != nil {
		return "", e
	}
	split := strings.Split(string(output), "\n")
	return strings.Split(split[1], " ")[0], nil

}

func TestNewDriverFromConfig(t *testing.T) {
	conf := `
url=123
username=user@abcde123213
password=123444
insecure=true
cafile=
ovirtVmName=kube-node-1

`
	ovirt, e := newDriver(strings.NewReader(conf))
	if e != nil {
		t.Error(e)
	}
	if ovirt.Connection.Url != "123" {
		t.Errorf("failed parsing url")
	}
	if ovirt.Connection.Username != "user@abcde123213" {
		t.Errorf("failed parsing username")
	}
	if ovirt.Connection.Password != "123444" {
		t.Errorf("failed parsing password")
	}
	if ovirt.Connection.Insecure != true {
		t.Errorf("failed parsing insecure")
	}
	if ovirt.Connection.CAFile != "" {
		t.Errorf("failed parsing cafile")
	}
	if ovirtVmName != "kube-node-1" {
		t.Errorf("failed parsing ovirtVmName")
	}
}
