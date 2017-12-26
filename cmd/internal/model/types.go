package model

type DiskAttachment struct {
	Id          string `json:"id"`
	Bootable    bool   `json:"bootable"`
	PassDiscard bool   `json:"pass_discard"`
	Interface   string `json:"interface"`
	Active      bool   `json:"active"`
	Disk        Disk   `json:"disk"`
}

type Disk struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	ProvisionedSize string `json:"provisioned_size`
}
