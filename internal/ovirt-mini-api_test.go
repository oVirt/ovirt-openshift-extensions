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
	"errors"
	"fmt"
	"gopkg.in/gcfg.v1"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var driverConfig string

func init() {
	driverConfig = os.Getenv("OVIRT_FLEXDRIVER_CONF")
}

func TestLoadConf(t *testing.T) {
	api := Ovirt{}
	err := gcfg.ReadFileInto(&api, driverConfig)
	if err != nil {
		t.Fatal(err)
	}
	// sanity check the config is loaded
	if api.Connection.Url == "" {
		t.Fatal("empty connection url")
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
	// create test server with handler
	api := prepareApi(tokenHandlerFunc(200))

	err := api.Authenticate()

	if err != nil {
		t.Fatalf("failed authentication %s", err)
	}

	if api.token.ExpireIn != 200 {
		t.Fatalf("token expiration expected: 200, got: %v", api.token.ExpireIn)
	}

	if api.token.ExpirationTime.Before(time.Now()) {
		t.Fatalf("token should expire only within 200 sec, but expires on %s", api.token.ExpirationTime)
	}
}

func TestAttach(t *testing.T) {
	expectedId := 1234
	api := prepareApi(func(w http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(w, `{ "id": "%v", "bootable": "true", "pass_discard": "true", "interface": "ide", "active":"true"}`, expectedId)
	})

	response, e := api.Attach(
		AttachRequest{"data", "vm_disk_1", "ext4", "rw", "", ""},
		"host1")

	if e != nil {
		t.Fatal(e)
	}

	if response.Status != Success {
		t.Error(errors.New("response failure"))
	}

	if response.Device != "/dev/disk/by-id/virtio"+string(expectedId) {
		t.Error(errors.New("different device paths"))
	}

	t.Log(response)
}

func prepareApi(handler http.HandlerFunc) Ovirt {
	ts := httptest.NewServer(handler)
	api := getApi(http.DefaultClient)
	api.Connection.Url = ts.URL
	api.Connection.Insecure = true
	return api
}

func tokenHandlerFunc(expireIn int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{ "access_token": "1234567890", "exp": "%v", "token_type": "Bearer"}`, expireIn)
	}
}

func getApi(client *http.Client) Ovirt {
	return Ovirt{}
}
