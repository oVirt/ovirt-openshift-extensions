package main

import (
	"fmt"
	"io"
	"errors"
	"gopkg.in/gcfg.v1"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"

	"context"
	"github.com/ovirt/ovirt-openshift-extensions/internal"
	"k8s.io/api/core/v1"
)

// ProviderName is the canonical name the plugin will register under. It must be different the the in-tree
// implementation name, "ovirt". The addition of "ecp" stand for External-Cloud-Provider
const ProviderName = "ovirt-cloud-provider"
const DefaultVMSearchQuery = "vms?follow=nics&search="

type OvirtNode struct {
	UUID      string
	Name      string
	IPAddress string
}

type ProviderConfig struct {
	Filters struct {
		VmsQuery string `gcfg:"vmsquery"`
	}
}

type CloudProvider struct {
	VmsQuery string
	internal.OvirtApi
}

// init will register the cloud provider
func init() {
	glog.Info("about to register the ovirt cloud provider to the cluster")
	cloudprovider.RegisterCloudProvider(
		ProviderName,
		func(config io.Reader) (cloudprovider.Interface, error) {
			if config == nil {
				return nil, fmt.Errorf("missing configuration file for ovirt cloud provider")
			}
			ovirtClient, err := internal.NewOvirt(config)
			if err != nil {
				return nil, err
			}

			providerConfig := ProviderConfig{}
			err = gcfg.ReadInto(&providerConfig, config)
			if err != nil {
				return nil, err
			}
			return NewOvirtProvider(&providerConfig, ovirtClient)
		})
}

func NewOvirtProvider(providerConfig *ProviderConfig, ovirtApi internal.OvirtApi) (*CloudProvider, error) {
	// TODO consider some basic validations for the search query although it can be tricky
	if ovirtApi.GetConnectionDetails().Url == "" {
		return nil, errors.New("oVirt engine url is empty")
	}

	vmsQuery := DefaultVMSearchQuery + providerConfig.Filters.VmsQuery
	return &CloudProvider{vmsQuery, ovirtApi}, nil

}

// Initialize provides the cloud with a kubernetes client builder and may spawn goroutines
// to perform housekeeping activities within the cloud provider.
func (p *CloudProvider) Initialize(clientBuilder controller.ControllerClientBuilder) {
	p.Authenticate()
}

// LoadBalancer returns a balancer interface. Also returns true if the interface is supported, false otherwise.
func (*CloudProvider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return nil, false
}

// Instances returns an instances interface. Also returns true if the interface is supported, false otherwise.
func (p *CloudProvider) Instances() (cloudprovider.Instances, bool) {
	return p, true
}

// Zones returns a zones interface. Also returns true if the interface is supported, false otherwise.
func (*CloudProvider) Zones() (cloudprovider.Zones, bool) {
	return nil, false

}

// NodeAddressses returns an hostnames/external-ips of the calling node
// TODO how to detect a primary external IP? how to pass hostnames if we have it?
func (p *CloudProvider) NodeAddresses(context context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	vms, err := p.getVms()
	if err != nil {
		return nil, err
	}

	var vm = vms[string(name)]
	if vm.Id == "" {
		return nil, fmt.Errorf(
			"VM by the name %s does not exist."+
				" The VM may have been removed, or the search query criteria needs correction",
			name)
	}

	// TODO the old provider supplied hostnames - look for fqdn of VM maybe?. Consider implementing.
	addresses := extractNodeAddresses(vm)
	return addresses, nil
}

// InstanceID returns the ovirt VM id by the vm name which is nodeName.
// Note that if the VM does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (p *CloudProvider) InstanceID(context context.Context, nodeName types.NodeName) (string, error) {
	vms, err := p.getVms()
	vm, ok := vms[string(nodeName)]
	if !ok || vm.Status == "down" {
		return "", cloudprovider.InstanceNotFound
	}
	return vm.Id, err
}

// Clusters returns a clusters interface.  Also returns true if the interface is supported, false otherwise.
func (*CloudProvider) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

// Routes returns a routes interface along with whether the interface is supported.
func (*CloudProvider) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

// ProviderName returns the cloud provider ID.
func (*CloudProvider) ProviderName() string {
	return ProviderName
}

// ScrubDNS provides an opportunity for cloud-provider-specific code to process DNS settings for pods.
// TODO to be removed - unsed and deprecated in latest cloud provider API
// See https://github.com/kubernetes/kubernetes/commit/65efeee64f772e0f38037e91a677138a335a7570#diff-449f7e867e01d9cffec053fdd28b6ffc
func (*CloudProvider) ScrubDNS(nameservers, searches []string) (nsOut, srchOut []string) {
	return nil, nil

}

// HasClusterID returns true if a ClusterID is required and set
func (*CloudProvider) HasClusterID() bool {
	return false
}

//AddSSHKeyToAllInstances Not implemented in ovirt - Can be implemented by pushing the keys to
// cloud-init and rebooting the VM. Don't know if that's needed when bootstraping a node
func (*CloudProvider) AddSSHKeyToAllInstances(context context.Context, user string, keyData []byte) error {
	return errors.New("NotImplemented")
}

func (p *CloudProvider) CurrentNodeName(context context.Context, hostname string) (types.NodeName, error) {
	vm, err := p.GetVM(hostname)
	return types.NodeName(vm.Fqdn), err
}

// ExternalID returns the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (p *CloudProvider) ExternalID(nodeName types.NodeName) (string, error) {
	vms, err := p.getVms()
	if err != nil  {
		return "", err
	}
	vm, ok :=  vms[string(nodeName)]
	if !ok {
		return "", cloudprovider.InstanceNotFound
	}
	return vm.Id, nil
}

// InstanceExistsByProviderID returns true if the instance for the given provider id still is running.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
func (p *CloudProvider) InstanceExistsByProviderID(context context.Context, providerID string) (bool, error) {
	vms, err := p.getVms()
	if err != nil {

	}
	for _, v := range vms {
		// statuses of up, unknown, not-responding still most likely indicate a running
		// instance. First lets consider 'down' as the non existing instance.
		if v.Id == providerID {
			if v.Status == "down" {
				return false, nil
			} else {
				return true, nil
			}
		}
	}
	return false, fmt.Errorf("there is no instance with ID %s", providerID)
}

func (p *CloudProvider) InstanceShutdownByProviderID(context context.Context, providerID string) (bool, error) {
	vmsById, e := p.getVmsById()
	if e != nil {
		return false, e
	}

	vm, ok := vmsById[providerID]

	if  !ok {
		return false, fmt.Errorf("vm with id %s doesn't exist", providerID)
	}

	return vm.Status == "down", nil
}

// InstanceType returns the type of the specified instance.
func (p *CloudProvider) InstanceType(context context.Context, name types.NodeName) (string, error) {
	// TODO unclear what this method is needed for
	return "", cloudprovider.NotImplemented
}

// InstanceTypeByProviderID returns the type of the specified instance.
func (p *CloudProvider) InstanceTypeByProviderID(context context.Context, providerID string) (string, error) {
	return "", cloudprovider.NotImplemented
}

func (p *CloudProvider) NodeAddressesByProviderID(context context.Context, providerID string) ([]v1.NodeAddress, error) {
	vmsById, err := p.getVmsById()
	if err != nil {
		return nil, err
	}

	vm, ok := vmsById[string(providerID)]
	if !ok {
		return nil, fmt.Errorf(
			"VM by the ID %s does not exist."+
				" The VM may have been removed, or the search query criteria needs correction",
			providerID)
	}

	// TODO the old provider supplied hostnames - look for fqdn of VM maybe?. Consider implementing.
	addresses := extractNodeAddresses(vm)
	return addresses, nil
}

func (p *CloudProvider) getVms() (map[string]internal.VM, error) {
	vms, err := p.GetVMs(p.VmsQuery)

	var vmsMap = make(map[string]internal.VM, len(vms))
	for _, v := range vms {
		vmsMap[v.Name] = v
	}

	return vmsMap, err
}

func (p *CloudProvider) getVmsById() (map[string]internal.VM, error) {
	vms, e := p.getVms()
	if e != nil {
		return vms, e
	}
	vmsById := make(map[string]internal.VM, len(vms))
	for _, vm := range vms {
		vmsById[vm.Id] = vm
	}
	return vmsById, nil
}

// extractNodeAddresses will return all addresses of the reported node
// TODO how to detect a primary external IP? how to pass hostnames if we have it?
func extractNodeAddresses(vm internal.VM) []v1.NodeAddress {
	addresses := make([]v1.NodeAddress,0)
	for _, nics := range vm.Nics.Nics {
		for _, dev := range nics.Devices.Devices {
			for _, ip := range dev.Ips.Ips {
				addresses = append(addresses, v1.NodeAddress{Address: ip.Address, Type:v1.NodeExternalIP})
			}
		}
	}
	return addresses
}
