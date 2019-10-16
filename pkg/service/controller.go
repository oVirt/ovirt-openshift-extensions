package service

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"

	"github.com/ovirt/ovirt-openshift-extensions/internal"
	"github.com/ovirt/ovirt-openshift-extensions/pkg/common"
)

type ControllerService struct {
	ovirtApi internal.OvirtApi
}

var ControllerCaps = []csi.ControllerServiceCapability_RPC_Type{
	csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME, // attach/detach
}

func (c *ControllerService) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	// idempotence first - see if disk already exists, ovirt creates disk by name(alias in ovirt as well)
	diskByName, e := c.ovirtApi.GetDiskByName(req.Name)
	if e != nil {
		klog.Error(e)
		return nil, e
	}
	// if exists we're done
	if len(diskByName.Disks) == 1 {
		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				CapacityBytes:      int64(diskByName.Disks[0].ProvisionedSize),
				VolumeId:           diskByName.Disks[0].Id,
				VolumeContext:      nil,
				ContentSource:      nil,
				AccessibleTopology: nil,
			},
		}, nil
	}

	// creating the disk
	disk, e := c.ovirtApi.CreateUnattachedDisk(
		req.Name,
		req.Parameters[common.ParameterStorageDomainName],
		req.CapacityRange.GetRequiredBytes(),
		false,
		common.ParseBoolParam(req.Parameters[common.ParameterDiskThinProvisioning], true),
	)
	if e != nil {
		return nil, e
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			CapacityBytes: int64(disk.ProvisionedSize),
			VolumeId:      disk.Id,
		},
	}, nil
}

func (c *ControllerService) DeleteVolume(ctx context.Context,req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	// idempotence first - see if disk already exists, ovirt creates disk by name(alias in ovirt as well)
	_, e := c.ovirtApi.Get("disks/" + req.VolumeId)
	if e != nil {
		if _, ok := e.(internal.NotFound); ok {
			// the disk is not there, we're done
			return &csi.DeleteVolumeResponse{}, nil
		}
		klog.Error(e)
		return nil, e
	}
	// disks exists lets remove it
	_, e = c.ovirtApi.Delete("disks/" + req.VolumeId)
	if e != nil {
		return nil, e
	}
	return &csi.DeleteVolumeResponse{}, nil
}

func (c *ControllerService) ControllerPublishVolume(context.Context, *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *ControllerService) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *ControllerService) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *ControllerService) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *ControllerService) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *ControllerService) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *ControllerService) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *ControllerService) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *ControllerService) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *ControllerService) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	caps := make([]*csi.ControllerServiceCapability, 0, len(ControllerCaps))
	for _, c := range ControllerCaps {
		caps = append(
			caps,
			&csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: c,
					},
				},
			},
		)
	}
	return &csi.ControllerGetCapabilitiesResponse{Capabilities: caps}, nil
}
