package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"k8s.io/klog"

	"github.com/ovirt/ovirt-openshift-extensions/internal"
	"github.com/ovirt/ovirt-openshift-extensions/pkg/service"
)

var (
	endpoint            = flag.String("endpoint", "unix:/tmp/csi.sock", "CSI endpoint")
	ovirtConfigFilePath = flag.String("ovirt-conf", "", "Path to ovirt api config")
)

func init() {
	flag.Set("logtostderr", "true")
	klog.InitFlags(flag.CommandLine)
}


func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	handle()
	os.Exit(0)
}

func handle() {
	if service.VendorVersion == "" {
		klog.Fatalf("VendorVersion must be set at compile time")
	}
	klog.V(2).Infof("Driver vendor %v %v", service.VendorName, service.VendorVersion)

	ovirtApi, err := newOvirt()
	if err != nil {
		klog.Fatalf("Failed to initialize ovirt client %s", err)
	}

    driver := service.NewOvirtCSIDriver(ovirtApi)

    driver.Run(*endpoint)
}

func newOvirt() (internal.OvirtApi, error) {
	var conf string
	value, exist := os.LookupEnv("OVIRT_API_CONF")
	if exist {
		conf = value
	} else {
		conf = *ovirtConfigFilePath
	}
	file, e := os.Open(conf)
	if e != nil {
		return nil, e
	}
	ovirt, err := internal.NewOvirt(file)
	if err != nil {
		return nil, err
	}
	err = ovirt.Authenticate()
	if err != nil {
		return nil, err
	}
	return ovirt, nil
}