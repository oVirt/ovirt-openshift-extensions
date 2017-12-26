package internal

import "github.com/ovirt/ovirt-flexdriver/internal"

var apiCases = []struct {
	description string
	in          internal.AttachRequest
	want        internal.AttchResponse
}{
	{
		"attach a disk to vm, positive",
		internal.AttachRequest{
			FsType:        "ext3",
			Mode:          "ro",
			StorageDomain: "domainId",
			VolumeName:    "domainName",
		},
		internal.AttchResponse{
			internal.Response{
				"success",
				"attached to vm successfully",
				"5a7b2687-07aa-4a7a-b589-4d4f847b9c29",
				"vol1",
				"attached",
				internal.Capabilities{""},
			},
		},
	},
}
