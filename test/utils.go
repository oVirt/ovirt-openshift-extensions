package test

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
)

type MockIdentityService struct {
	csi.IdentityServer
}
type MockControllerService struct {
	csi.ControllerServer
}
type MockNodeService struct {
	csi.NodeServer
}

