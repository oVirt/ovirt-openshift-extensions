package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ovirt/ovirt-openshift-extensions/internal"
	"encoding/json"
	"k8s.io/apimachinery/pkg/types"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

var testOvirtConfig = internal.Connection{
	"https: //fqdn:8443/ovirt-engine/api",
	"foo@domain",
	"123",
	true,
	"/dev/null",
}

var vmsJson string

const vm1NodeName = "master0.example.com"
const vm1Id = "f5fb9df5-19da-4d35-a0c8-f5d7569faacd"
const vm2Name = "centos"
const vm2Id = "f85501aa-afbb-46f2-a8d3-3dc299c07fee"

func init() {
	parse, err := ioutil.ReadFile("./vms.json")
	if err != nil {
		panic(err)
	}
	vmsJson = string(parse)
}

var _ = Describe("ovirt-cloud-provider configuration tests", func() {
	var (
		underTest *CloudProvider
		err	error
	)

	BeforeEach(func() {
		underTest, _ = NewOvirtProvider(&ProviderConfig{}, MockApi{testOvirtConfig})
	})

	Context("With a default config", func() {
		It("VMs query should have a default search vms " + DefaultVMSearchQuery, func() {
			Expect(underTest).ToNot(Equal(nil))
			Expect(underTest.VmsQuery).To(Equal(DefaultVMSearchQuery))
		})

		It("Return the cloud provider name", func() {
			Expect(underTest.ProviderName()).To(Equal(ProviderName))
		})

	})

	Context("With invalid config", func() {
		BeforeEach(func() {
			conf := ProviderConfig{}
			ovirtClient := MockApi{}
			underTest, err = NewOvirtProvider(&conf, ovirtClient)
		})
		It("should fail to start", func() {
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("ovirt-cloud-provider node tests", func() {

	var (
		underTest *CloudProvider
	)

	BeforeEach(func() {
		underTest, _ = NewOvirtProvider(&ProviderConfig{}, MockApi{internal.Connection{Url: "http://foo"}})
	})

	Context("With a node that exist on ovirt", func() {
		It("reports the instance exists", func() {
			exists, _ := underTest.InstanceExistsByProviderID(nil, vm1Id)
			Expect(exists).To(BeTrue())
		})

		It("returns the VM ID for for running VM by name", func() {
			id, _ := underTest.InstanceID(nil, vm1NodeName)
			Expect(id).To(Equal(vm1Id))
		})

		It("fails with InstanceNotFound when calling InstanceID for a down VM", func() {
			id, err := underTest.InstanceID(nil,vm2Name )
			Expect(err).To(Equal(cloudprovider.InstanceNotFound))
			Expect(id).To(Equal(""))
		})

		It("returns the current nodename VM FQDN", func() {
			nodeName, err := underTest.CurrentNodeName(nil, vm1NodeName)
			Expect(err).ToNot(HaveOccurred())
			Expect(nodeName).To(Equal(types.NodeName("etcd.example.com")))

		})

		It("returns a list of addresses for as reported by ovirt for the instance", func() {
			addresses, err := underTest.NodeAddresses(nil, vm1NodeName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(addresses).Should(HaveLen(3))
		})

		It("returns an empty list of addresses for a node which don't have interfaces", func() {
			addresses, err := underTest.NodeAddresses(nil, "dtestiso")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(addresses).Should(HaveLen(0))
		})

		It("returns an error for non-existing node", func() {
			_, err := underTest.NodeAddresses(nil, "nonexistingvm")
			Expect(err).Should(HaveOccurred())
		})

		It("returns true when instance with status up exists", func() {
			exists, err := underTest.InstanceExistsByProviderID(nil, vm1Id)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("returns true when instance is shutdown at oVirt", func() {
			exists, err := underTest.InstanceShutdownByProviderID(nil, "f85501aa-afbb-46f2-a8d3-3dc299c07fee")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("returns false when instance is up at oVirt", func() {
			exists, err := underTest.InstanceShutdownByProviderID(nil, vm1Id)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("returns false when instance is unknown at oVirt", func() {
			exists, err := underTest.InstanceShutdownByProviderID(nil, "f9772ed4-2ee4-4386-8a8a-22e3b73add68")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("returns node address by ID", func() {
			addresses, err := underTest.NodeAddressesByProviderID(nil, vm1Id)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(addresses).Should(HaveLen(3))
		})

		It("returns the VM ID for the node name (node name should be the vm name)", func() {
			vmId, err := underTest.ExternalID(vm1NodeName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(vmId).To(Equal(vm1Id))
		})

		It("fails when a node ExternalID call can't find the node(node name should be the vm name)", func() {
			vmId, err := underTest.ExternalID("non-existing-VM")
			Expect(err).Should(Equal(cloudprovider.InstanceNotFound))
			Expect(vmId).To(Equal(""))
		})

	})
})

var _ = Describe("ovirt-cloud-provider ", func() {

	var (
		underTest *CloudProvider
	)

	BeforeEach(func() {
		underTest, _ = NewOvirtProvider(&ProviderConfig{}, MockApi{internal.Connection{Url:"http://foo"}})
	})

	Context("example", func() {

		It("reports ", func() {
			exists, _ := underTest.InstanceExistsByProviderID(nil, vm1Id)
			Expect(exists).To(BeTrue())
		})

	})
})



type MockApi struct {
	Connection internal.Connection
}

func (MockApi) Authenticate() error {
	panic("implement me")
}

func (MockApi) Get(path string) ([]byte, error) {
	panic("implement me")
}

func (MockApi) Post(path string, data interface{}) (string, error) {
	panic("implement me")
}

func (MockApi) Delete(path string) ([]byte, error) {
	panic("implement me")
}

func (m MockApi) GetVM(name string) (internal.VM, error) {
	vms, err := m.GetVMs("")
	vmsMap := make(map[string]internal.VM, len(vms))
	for _,v := range vms {
		vmsMap[v.Name] = v
	}
	return vmsMap[name], err
}
func (MockApi) GetVMs(query string) ([]internal.VM, error) {
	vmResult := internal.VMResult{}
	err := json.Unmarshal([]byte(vmsJson), &vmResult)
	return vmResult.Vms, err
}

func (MockApi) GetDiskAttachment(vmId, diskId string) (internal.DiskAttachment, error) {
	panic("implement me")
}

func (MockApi) GetDiskAttachments(vmId string) ([]internal.DiskAttachment, error) {
	panic("implement me")
}

func (MockApi) DetachDiskFromVM(vmId string, diskId string) error {
	panic("implement me")
}

func (MockApi) Attach(params internal.AttachRequest, nodeName string) (internal.Response, error) {
	panic("implement me")
}

func (MockApi) GetDiskByName(diskName string) (internal.DiskResult, error) {
	panic("implement me")
}

func (MockApi) CreateUnattachedDisk(diskName string, storageDomainName string, sizeIbBytes int64, readOnly bool, diskFormat string) (internal.Disk, error) {
	panic("implement me")
}

func (MockApi) CreateDisk(
	diskName string,
	storageDomainName string,
	readOnly bool,
	vmId string,
	diskId string,
	diskInterface string) (internal.DiskAttachment, error) {
	panic("implement me")
}

func (m MockApi) GetConnectionDetails() internal.Connection {
	return m.Connection

}