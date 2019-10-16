package common

import (
	"strconv"

	"k8s.io/klog"
)

const (
	ParameterStorageDomainName = "ovirtStorageDomain"
	ParameterDiskThinProvisioning = "ovirtDiskThinProvisioning"
	ParameterFsType = "fsType"
)

func ParseBoolParam(value string, defaultValue bool) bool {
	b, e := strconv.ParseBool(value)
		if e != nil {
			klog.Warningf("wrong value for while converting %s to boolean. Using default %v", value, defaultValue)
			return defaultValue
		} else {
			return b
		}
}