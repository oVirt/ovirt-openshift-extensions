package main

import "testing"

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
