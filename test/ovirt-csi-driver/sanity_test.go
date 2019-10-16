package ovirt_csi_driver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	sanity "github.com/kubernetes-csi/csi-test/pkg/sanity"

	"github.com/ovirt/ovirt-openshift-extensions/internal"
	driver "github.com/ovirt/ovirt-openshift-extensions/pkg/service"
)

func TestSanity(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "ovirt-csi-tests-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	endpoint := fmt.Sprintf("unix:%s/csi.sock", tmpDir)
	mountPath := path.Join(tmpDir, "mount")
	stagePath := path.Join(tmpDir, "stage")

	ovirtMock := internal.NewMockOvirt()

	driver.VendorVersion = driver.VendorVersion + "-sanity-tests"
	driver := driver.NewOvirtCSIDriver(ovirtMock)

	//instance := &compute.Instance{
	//	Name:  "test-name",
	//	Disks: []*compute.AttachedDisk{},
	//}
	//cloudProvider.InsertInstance(instance, "test-location", "test-name")

	go func() {
		driver.Run(endpoint)
	}()

	// Run test
	config := &sanity.Config{
		TargetPath:  mountPath,
		StagingPath: stagePath,
		Address:     endpoint,
	}
	sanity.Test(t, config)
}
