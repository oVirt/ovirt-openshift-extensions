package service

import (
	"k8s.io/klog"

	"github.com/ovirt/ovirt-openshift-extensions/internal"
)

var (
	// set by ldflags
	VendorVersion string
	VendorName    = "csi.ovirt.org"
)

type OvirtCSIDriver struct {
	*IdentityService
	*ControllerService
	*NodeService
	ovirtApi internal.OvirtApi
}

func NewOvirtCSIDriver(ovirtApi internal.OvirtApi) *OvirtCSIDriver {
	d := OvirtCSIDriver{
		IdentityService:   &IdentityService{ovirtApi},
		ControllerService: &ControllerService{ovirtApi},
		NodeService:       &NodeService{ovirtApi},
		ovirtApi:          ovirtApi,
	}
	return &d
}

func (driver *OvirtCSIDriver) Run(endpoint string) {
	// run the gRPC server
	klog.Info("Setting the rpc server")

    s := NewNonBlockingGRPCServer()
    s.Start(endpoint, driver.IdentityService, driver.ControllerService, driver.NodeService)
    s.Wait()
}
