package service

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"

	"github.com/ovirt/ovirt-openshift-extensions/internal"
)

type NodeService struct {
	ovirtApi internal.OvirtApi
}

var NodeCaps = []csi.NodeServiceCapability_RPC_Type{
	csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
}

func (n *NodeService) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	panic("implement me")
}

func (n *NodeService) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	panic("implement me")
}

func (n *NodeService) NodePublishVolume(context.Context, *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	panic("implement me")
}

func (n *NodeService) NodeUnpublishVolume(context.Context, *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	panic("implement me")
}

func (n *NodeService) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	panic("implement me")
}

func (n *NodeService) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	panic("implement me")
}

func (n *NodeService) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	panic("implement me")
}

func (n *NodeService) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	caps := make([]*csi.NodeServiceCapability, 0, len(NodeCaps))
	for _, c := range NodeCaps {
		caps = append(
			caps,
			&csi.NodeServiceCapability{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: c,
					},
				},
			},
		)
	}
	return &csi.NodeGetCapabilitiesResponse{Capabilities: caps}, nil
}
