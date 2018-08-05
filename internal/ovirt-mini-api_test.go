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
	"code.cloudfoundry.org/bytefmt"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	if ovirt.Connection.Url != "123" {
		t.Errorf("failed parsing url")
	}
	if ovirt.Connection.Username != "user@abcde123213" {
		t.Errorf("failed parsing username")
	}
	if ovirt.Connection.Password != "123444" {
		t.Errorf("failed parsing password")
	}
	if ovirt.Connection.Insecure != true {
		t.Errorf("failed parsing insecure")
	}
	if ovirt.Connection.CAFile != "" {
		t.Errorf("failed parsing cafile")
	}
}

// TestAuthenticateWithUnexpiredToken makes sure we reuse the auth token
func TestAuthenticateWithUnexpiredToken(t *testing.T) {
	api := prepareApi(tokenHandlerFunc(10000000))
	err := api.Authenticate()
	if err != nil {
		t.Fatalf("failed authentication %s", err)
	}
}

func TestFetchToken(t *testing.T) {
	// ignore loading the token
	tokenStore = "/dev/null"
	// expire in 1 month from now
	expiredIn := time.Now().AddDate(0, 1, 0).UnixNano()
	// create test server with handler
	api := prepareApi(tokenHandlerFunc(expiredIn))

	err := api.Authenticate()

	if err != nil {
		t.Fatalf("failed authentication %s", err)
	}

	if api.token.ExpireIn != expiredIn {
		t.Fatalf("token expiration expected: %v, got: %v", expiredIn, api.token.ExpireIn)
	}

	if api.token.ExpirationTime.Before(time.Now()) {
		t.Fatalf("token should expire in the future, but expired on on %v ", api.token.ExpirationTime)
	}
}

func prepareApi(handler http.HandlerFunc) Ovirt {
	ts := httptest.NewServer(handler)
	api := getApi(http.DefaultClient)
	api.Connection.Url = ts.URL
	api.Connection.Insecure = true
	return api
}

func tokenHandlerFunc(expireIn int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{ "access_token": "1234567890", "exp": "%v", "token_type": "Bearer"}`, expireIn)
	}
}

func getApi(client *http.Client) Ovirt {
	return Ovirt{}
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
	api := prepareApi(genericRequestHandlerFunc(createResponse))
	_, e := api.CreateUnattachedDisk(
		"pvc-d69b93df-7e96-11e8-b3fa-001a4a160100",
		"iscidomain",
		1073741824,
		false,
		"cow")
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
	api := prepareApi(genericRequestHandlerFunc(attachResponse))
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
	api := prepareApi(genericRequestHandlerFunc(detachResponse))
	e := api.DetachDiskFromVM(vmId, diskId)
	if e != nil {
		t.Errorf(e.Error())
	}
}
