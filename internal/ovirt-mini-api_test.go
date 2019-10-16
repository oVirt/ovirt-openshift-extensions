/*
Copyright 2017 oVirt-maintainers

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"code.cloudfoundry.org/bytefmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const vmId = "12345678-1234-1234-1234-123456789101"
const diskId = "d69b93df-7e96-11e8-b3fa-001a4a160100"

func TestLoadConf(t *testing.T) {
	conf := `
url=123
username=user@abcde123213
password=123444
insecure=true
cafile=
`
	ovirt, e := NewOvirt(strings.NewReader(conf))
	if e != nil {
		t.Error(e)
	}
	if ovirt.GetConnectionDetails().Url != "123" {
		t.Errorf("failed parsing url")
	}
	if ovirt.GetConnectionDetails().Username != "user@abcde123213" {
		t.Errorf("failed parsing username")
	}
	if ovirt.GetConnectionDetails().Password != "123444" {
		t.Errorf("failed parsing password")
	}
	if ovirt.GetConnectionDetails().Insecure != true {
		t.Errorf("failed parsing insecure")
	}
	if ovirt.GetConnectionDetails().CAFile != "" {
		t.Errorf("failed parsing cafile")
	}
}

var _ = Describe("Authentication tests", func() {

	Context("token test", func() {

		AfterEach(func() {
			os.Remove("/tmp/ovirt-flexdriver.token")
		})

		It("fetches a valid token", func() {
			api := CreateMockOvirtClient(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{ "access_token": "1234567890", "exp": "%v", "token_type": "Bearer"}`, 10000000)
			})
			err := api.Authenticate()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(api.token).NotTo(BeNil())
			Expect(api.token.ExpireIn).To(BeNumerically("==", 10000000))
		})

		It("persists the token to the token store", func() {
			api := CreateMockOvirtClient(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{ "access_token": "1234567890", "exp": "%v", "token_type": "Bearer"}`, 10000000)
			})
			err := api.Authenticate()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(api.token.Value).NotTo(Equal(""))

			_, err = os.Stat("/tmp/ovirt-flexdriver.token")
			Expect(err).ShouldNot(HaveOccurred())

		})

		It("correctly translate unix time to expiration time", func() {
			// expire in 1 month from now
			expiredIn := time.Now().AddDate(0, 1, 0).UnixNano()
			// create test server with handler
			api := CreateMockOvirtClient(tokenHandlerFunc(expiredIn))
			err := api.Authenticate()
			Expect(err).NotTo(HaveOccurred())
			Expect(api.token.ExpireIn).To(Equal(expiredIn))
			Expect(api.token.ExpirationTime.Month()).To(
				Equal(time.Now().AddDate(0, 1, 0).Month()))
		})

		It("fails when fetch token moved 302", func() {
			api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(302)
			})
			err := api.Authenticate()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when authenticating again", func() {
		Context("authenticating again when token is valid", func() {
			It("doesn't need to re authenticate with password", func() {})
			It("loads the existing token from the token store", func() {})
		})

		Context("old token loaded from the token store is expired", func() {
			It("fetches the token again", func() {})
			It("updates the token value", func() {})
			It("serializes the token to store", func() {})
			It("successfully authenticate", func() {})
		})

		Context("old token file doesn't exists", func() {
			It("authenticate using user and pass", func() {})
			It("stores the token to the store", func() {})
		})
	})

	Context("when using the token and error occurs", func() {
		Context("on 302 moved error", func() {
			It("returns a proper error", func() {})
		})

		Context("on 401 unauthorized", func() {
			It("catches it and reauthenticate", func() {})
		})
	})

})

var _ = Describe("API calls on resources", func() {

	Context("Create Unattached Disk, thick provisioning ", func() {
		Context("test the sent request", func() {
			var underTest []byte
			var errRead error
			api := NewMockOvirt()

			api.Handle("/disks", func(writer http.ResponseWriter, request *http.Request) {
				underTest, errRead = ioutil.ReadAll(request.Body)
			})

			_, _ = api.CreateUnattachedDisk(
				"disk1",
				"data1",
				19999,
				false,
				false)

			req := make(map[string]interface{})
			err := json.Unmarshal(underTest, &req)

			It("a valid json", func() {
				Expect(errRead).ShouldNot(HaveOccurred())
				Expect(err).NotTo(HaveOccurred())
			})

			It("has the disk name", func() {
				Expect(req["name"]).To(Equal("disk1"))
			})
			It("has disk type raw", func() {
				Expect(req["format"]).To(Equal("raw"))
			})
			Specify("sparse is false", func() {
				Expect(req["sparse"]).To(Equal("false"))
			})
		})
	})

	Context("Create Unattached Disk, thin provisioning (the default) isci", func() {
		Context("test the sent request", func() {
			var underTest []byte
			var errRead error
			api := NewMockOvirt()

			api.Handle("/storagedomains", func(writer http.ResponseWriter, request *http.Request) {
				writer.Write([]byte(`{"storage_domain": [{"name": "data1", "storage": {"type": "iscsi"}}] }`))
			})

			api.Handle("/disks", func(writer http.ResponseWriter, request *http.Request) {
				underTest, errRead = ioutil.ReadAll(request.Body)
			})

			_, _ = api.CreateUnattachedDisk(
				"disk1",
				"data1",
				19999,
				false,
				true)

			req := make(map[string]interface{})
			err := json.Unmarshal(underTest, &req)

			It("a valid json", func() {
				Expect(errRead).ShouldNot(HaveOccurred())
				Expect(err).NotTo(HaveOccurred())
			})

			It("must provide storage type if set ", func() {
				Expect(req["storage_domains"]).To(HaveKey("storage_domain", ))
				Expect(req["storage_domains"]).To(HaveLen(1))
			})
			It("has the disk name", func() {
				Expect(req["name"]).To(Equal("disk1"))
			})
			It("has disk type raw", func() {
				Expect(req["format"]).To(Equal("cow"))
			})
			Specify("sparse is false", func() {
				Expect(req["sparse"]).To(Equal("true"))
			})
		})
	})

	Context("Get Storage Domain by name that exists", func() {
		api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte(`{"storage_domain": [{"name":"foo", "storage": {"type":"iscsi"}}] }`))
		})
		d, err := api.GetStorageDomainBy("foo")

		It("returns with no error", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("returns a storage domain", func() {
			Expect(d).NotTo(BeZero())
		})
		It("has the right name", func() {
			Expect(d.Name).To(Equal("foo"))
		})
		It("has the right domain type", func() {
			Expect(d.Storage.Type).To(Equal("iscsi"))
		})
	})
	Context("Search Storage Domain by name that doesn't exist", func() {
		api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte(`{ }`))
		})
		d, err := api.GetStorageDomainBy("non-existing")

		It("returns error 'not exist'", func() {
			Expect(err).Should(Equal(ErrNotExist))
		})
		It("returns a zero-valued domain", func() {
			Expect(d).To(BeZero())
		})
		It("has no name", func() {
			Expect(d.Name).To(Equal(""))
		})
		It("has no storage type", func() {
			Expect(d.Storage.Type).To(Equal(""))
		})
	})
	Context("Default storage params", func() {
		Context("For thick provisioning", func() {
			api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
				writer.Write([]byte(`{ }`))
			})
			format, sparse, err := api.DefaultDiskParamsBy("", false)
			It("returns with no error", func() {
				Expect(err).ShouldNot(HaveOccurred())
			})
			Specify("disk format is raw", func() {
				Expect(format).To(Equal(DiskFormat("raw")))
			})
			Specify("not sparse", func() {
				Expect(sparse).To(Equal(Sparse(false)))
			})
		})
		Context("For thin provisioning", func() {
			Context("for block domain (iscsi/fc)", func() {
				api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
					writer.Write([]byte(`{ "storage_domain": [{"name": "data", "storage": {"type": "iscsi"}}]}`))
				})
				format, sparse, err := api.DefaultDiskParamsBy("data", true)
				It("returns with no error", func() {
					Expect(err).ShouldNot(HaveOccurred())
				})
				Specify("disk format is cow", func() {
					Expect(format).To(Equal(DiskFormat("cow")))
				})
				Specify("use sparse", func() {
					Expect(sparse).To(Equal(Sparse(true)))
				})
			})
			Context("for file domain nfs", func() {
				api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
					writer.Write([]byte(`{ "storage_domain": [{"name": "data", "storage": {"type": "gluster"}}]}`))
				})
				format, sparse, err := api.DefaultDiskParamsBy("data", true)
				It("returns with no error", func() {
					Expect(err).ShouldNot(HaveOccurred())
				})

				Specify("disk format is raw", func() {
					Expect(format).To(Equal(DiskFormat("raw")))
				})
				Specify("use sparse", func() {
					Expect(sparse).To(Equal(Sparse(true)))
				})
			})
			Context("for file domain gluster", func() {
				api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
					writer.Write([]byte(`{ "storage_domain": [{"name": "data", "storage": {"type": "nfs"}}]}`))
				})
				format, sparse, err := api.DefaultDiskParamsBy("data", true)
				It("returns with no error", func() {
					Expect(err).ShouldNot(HaveOccurred())
				})
				Specify("disk format is raw", func() {
					Expect(format).To(Equal(DiskFormat("raw")))
				})
				Specify("use sparse", func() {
					Expect(sparse).To(Equal(Sparse(true)))
				})
			})
		})
	})
	Context("VMs Get actions", func() {
		Context("Get Vm By Id", func() {
			Context("Vm exists", func() {
				api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
					writer.Write([]byte(`{ "name" : "centos",  "id": "f85501aa-afbb-46f2-a8d3-3dc299c07fee"}`))
				})
				vm, e := api.GetVMById("f85501aa-afbb-46f2-a8d3-3dc299c07fee")

				It("returns a VM instance", func() {
					Expect(vm.Id).To(Equal("f85501aa-afbb-46f2-a8d3-3dc299c07fee"))
				})

				It("returns with no error", func() {
					Expect(e).NotTo(HaveOccurred())
				})
			})
			Context("Vm does not exist", func() {
				api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
					http.NotFound(writer, request)
				})
				vm, e := api.GetVMById("f85501aa-afbb-46f2-a8d3-3dc299c07fee")

				It("returns a VM with no id", func() {
					Expect(vm.Id).To(Equal(""))
				})

				It("returns with a 404 error", func() {
					Expect(e).To(HaveOccurred())
					Expect(e).To(BeAssignableToTypeOf(NotFound{}))
				})
			})

		})
	})
})

func TestFailedFetchToken_move302(t *testing.T) {
	api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(302)
	})

	err := api.Authenticate()
	if err == nil {
		t.Fatal("should fail with error")
	}
}

func TestFailedFetchToken_404(t *testing.T) {
	api := CreateMockOvirtClient(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(404)
	})

	err := api.Authenticate()
	t.Logf("error is %s", err)
	if err == nil {
		t.Fatal("should fail with error")
	}
}

func tokenHandlerFunc(expireIn int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{ "access_token": "1234567890", "exp": "%v", "token_type": "Bearer"}`, expireIn)
	}
}

var testAttachRequest = `
{
	"capacity":"1G",
	"kubernetes.io/fsType":"ext4",
	"kubernetes.io/pvOrVolumeName":"test",
	"kubernetes.io/readwrite":"rw",
	"ovirtStorageDomain":"data1",
	"size":"1G",
	"volumeID":"disk123"
}
`

func TestAttachRequestFrom(t *testing.T) {
	request, e := AttachRequestFrom(testAttachRequest)
	if e != nil {
		t.Errorf(e.Error())
	}
	if request.Size != "1G" {
		t.Errorf("expected size is %v got %v", "1G", request.Size)
	}
}

func TestByteSizeFormatting(t *testing.T) {
	// ovirt api supports bytes. Lets expand with some literals

	// persistent volume style
	bytes, e := bytefmt.ToBytes("1G")
	if e != nil {
		t.Error(e)
	}
	t.Log(bytes)
}

func genericRequestHandlerFunc(json string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, json)
	}
}

func TestOvirt_CreateUnattachedDisk(t *testing.T) {
	createResponse := `
      {
        "id": "0138c56c-1937-461b-98e1-a1c5c82ae082",
        "name":"pvc-d69b93df-7e96-11e8-b3fa-001a4a160100",
        "actual_size":"0",
        "provisioned_size":"1073741824",
        "status": "locked",
        "format":"cow",
        "storage_domains": { "storage_domain": [{"name":"iscidomain"}] }
      }
    `
	api := CreateMockOvirtClient(genericRequestHandlerFunc(createResponse))
	_, e := api.CreateUnattachedDisk(
		"pvc-d69b93df-7e96-11e8-b3fa-001a4a160100",
		"iscidomain",
		1073741824,
		false,
		false)
	if e != nil {
		t.Errorf(e.Error())
	}
}

func TestOvirt_Attach(t *testing.T) {
	attachResponse := `
      {
        "id": "0138c56c-1937-461b-98e1-a1c5c82ae082",
        "name":"pvc-d69b93df-7e96-11e8-b3fa-001a4a160100",
        "actual_size":"0",
        "provisioned_size":"1073741824",
        "status": "locked",
        "format":"cow",
        "storage_domains": { "storage_domain": [{"name":"iscidomain"}] }
      }
    `
	api := CreateMockOvirtClient(genericRequestHandlerFunc(attachResponse))
	_, e := api.CreateDisk(
		"pvc-d69b93df-7e96-11e8-b3fa-001a4a160100",
		"iscidomain",
		false,
		"some-vm-id",
		"disk-uuid",
		"")
	if e != nil {
		t.Errorf(e.Error())
	}
}

func TestOvirt_DetachDiskFromVM(t *testing.T) {
	detachResponse := "{}"
	api := CreateMockOvirtClient(genericRequestHandlerFunc(detachResponse))
	e := api.DetachDiskFromVM(vmId, diskId)
	if e != nil {
		t.Errorf(e.Error())
	}
}
