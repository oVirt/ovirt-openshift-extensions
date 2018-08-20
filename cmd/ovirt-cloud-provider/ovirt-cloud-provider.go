package main

import (
	"fmt"
	"io"
	"path"

	"encoding/json"
	"errors"
	"github.com/go-ini/ini"
	"gopkg.in/gcfg.v1"
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	//"k8s.io/kubernetes/pkg/cloudprovider/providers/ovirt"
	"net"
	"unicode"

	"context"
	"github.com/ovirt/ovirt-openshift-extensions/internal"
	"k8s.io/api/core/v1"
	"os"
)

// ProviderName is the canonical name the plugin will register under. It must be different the the in-tree
// implementation name, "ovirt". The addition of "ecp" stand for External-Cloud-Provider
const ProviderName = "ovirt-ecp"

type OvirtNode struct {
	UUID      string
	Name      string
	IPAddress string
}

type ProviderConfig struct {
	Connection struct {
		Url      string `gcfg:"url"`
		Username string `gcfg:"username"`
		Password string `gcfg:"password"`
		Insecure bool   `gcfg:"insecure"`
		CAFile   string `gcfg:"cafile"`
	}
	Filters struct {
		VmsQuery string `gcfg:"vmsquery"`
	}
}

type CloudProvider struct {
	VmsQuery *url.URL
	OvirtApi *internal.OvirtApi
}

type VM struct {
	Name      string     `json:"name"`
	Id        string     `json:"id"`
	Fqdn      string     `json:"fqdn"`
	Addresses []net.Addr `json:""`
	Status    string     `json:"status"`
}

type VMs struct {
	Vm []VM
}

// init will register the cloud provider
func init() {
	cloudprovider.RegisterCloudProvider(
		ProviderName,
		func(config io.Reader) (cloudprovider.Interface, error) {
			if config == nil {
				return nil, fmt.Errorf("missing configuration file for ovirt cloud provider")
			}
			providerConfig := ProviderConfig{}
			err := gcfg.ReadInto(&providerConfig, config)
			if err != nil {
				return nil, err
			}
			return NewOvirtProvider(providerConfig)
		})
}

func NewOvirtProvider(providerConfig ProviderConfig) (*CloudProvider, error) {

	vmsQuery, err := url.Parse(providerConfig.Connection.Url)
	if err != nil {
		return nil, err
	}

	vmsQuery.Path = path.Join(vmsQuery.Path, "vms")
	vmsQuery.RawQuery = url.Values{"search": {providerConfig.Filters.VmsQuery}}.Encode()

	return &CloudProvider{VmsQuery: vmsQuery}, nil

}

func newOvirt() (*internal.Ovirt, error) {
	var conf string
	value, exist := os.LookupEnv("OVIRT_API_CONF")
	if exist {
		conf = value
	} else {
		conf = "/etc/ovirt/ovirt-api.conf"
	}

	cfg, err := ini.InsensitiveLoad(conf)
	if err != nil {
		return nil, err
	}
	connection := internal.Connection{}
	connection.Url = cfg.Section("").Key("url").String()
	connection.Username = cfg.Section("").Key("username").String()
	connection.Password = cfg.Section("").Key("password").String()
	connection.Insecure = cfg.Section("").Key("insecure").MustBool()
	connection.CAFile = cfg.Section("").Key("cafile").String()

	ovirt := internal.Ovirt{}
	ovirt.Connection = connection
	err = ovirt.Authenticate()
	if err != nil {
		return nil, err
	}
	return &ovirt, nil
}

// Initialize provides the cloud with a kubernetes client builder and may spawn goroutines
// to perform housekeeping activities within the cloud provider.
func (*CloudProvider) Initialize(clientBuilder controller.ControllerClientBuilder) {

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
func (p *CloudProvider) NodeAddresses(context context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	vms, err := p.getVms()
	if err == nil {
		return nil, err
	}

	var vm VM = vms[string(name)]
	if vm.Id == "" {
		return nil, fmt.Errorf(
			"VM by the name %s does not exist."+
				" The VM may have been removed, or the search query criteria needs correction",
			name)
	}
	if vm.Addresses == nil || len(vm.Addresses) == 0 {
		return nil, fmt.Errorf("Missing addresses of instance \n")
	}

	addresses := make([]v1.NodeAddress, len(vm.Addresses))

	for i, a := range vm.Addresses {
		var t v1.NodeAddressType
		if unicode.IsDigit(rune(a.String()[0])) {
			t = v1.NodeExternalIP
		} else {
			t = v1.NodeHostName
		}
		addresses[i] = v1.NodeAddress{
			Address: a.String(),
			Type:    t,
		}
	}
	return addresses, nil
}

func (p *CloudProvider) InstanceID(context context.Context, nodeName types.NodeName) (string, error) {
	vms, err := p.getVms()
	return vms[string(nodeName)].Id, err
}

func (p *CloudProvider) getVms() (map[string]VM, error) {
	resp, err := http.Get(p.VmsQuery.String())
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	vms := VMs{}
	err = json.NewDecoder(resp.Body).Decode(&vms)
	if err != nil {
		return nil, err
	}
	var vmsMap = make(map[string]VM)
	for i := 0; i < len(vms.Vm); i++ {
		v := vms.Vm[i]
		vmsMap[v.Name] = v
	}

	return vmsMap, nil
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
func (*CloudProvider) ScrubDNS(nameservers, searches []string) (nsOut, srchOut []string) {
	return nil, nil

}

// HasClusterID returns true if a ClusterID is required and set
func (*CloudProvider) HasClusterID() bool {
	return false
}

func (*CloudProvider) AddSSHKeyToAllInstances(context context.Context, user string, keyData []byte) error {
	return errors.New("NotImplemented")
}

func (*CloudProvider) CurrentNodeName(context context.Context, hostname string) (types.NodeName, error) {
	//var r types.NodeName = ""
	return types.NodeName(hostname), nil
}

// ExternalID returns the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (p *CloudProvider) ExternalID(nodeName types.NodeName) (string, error) {
	vms, err := p.getVms()
	if err != nil || vms[string(nodeName)].Id == "" {

	}
	return vms[string(nodeName)].Id, nil
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
	// TODO implement
	return false, nil
}

// InstanceType returns the type of the specified instance.
func (p *CloudProvider) InstanceType(context context.Context, name types.NodeName) (string, error) {
	return ProviderName, nil
}

// InstanceTypeByProviderID returns the type of the specified instance.
func (p *CloudProvider) InstanceTypeByProviderID(context context.Context, providerID string) (string, error) {
	return "", cloudprovider.NotImplemented
}

func (p *CloudProvider) NodeAddressesByProviderID(context context.Context, providerID string) ([]v1.NodeAddress, error) {
	return []v1.NodeAddress{}, cloudprovider.NotImplemented
}
