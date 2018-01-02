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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const caUrl = "ovirt-engine/services/pki-resource?resource=ca-certificate&format=X509-PEM-CA"
const tokenUrl = "/ovirt-engine/sso/oauth/token"
const tokenPayload = "grant_type=password&scope=ovirt-app-api&username=%s&password=%s"
const tokenStore = "/var/tmp/ovirt-flexdriver.token"

type Ovirt struct {
	Connection Connection
	client     http.Client
	token      Token
	api        OvirtApi
}

type Connection struct {
	Url      string `gcfg:"url"`
	Username string `gcfg:"username"`
	Password string `gcfg:"password"`
	Insecure bool   `gcfg:"insecure"`
	CAFile   string `gcfg:"cafile"`
}

type Token struct {
	Value          string `json:"access_token"`
	ExpireIn       int64  `json:"exp,string"`
	Type           string `json:"token_type"`
	ExpirationTime time.Time
}

func (ovirt *Ovirt) Authenticate() error {
	ovirtEngineUrl, err := url.Parse(ovirt.Connection.Url)
	if err != nil {
		return err
	}

	if ovirt.Connection.Insecure || ovirtEngineUrl.Scheme == "http" {
		ovirt.client = http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		// fetch ca if its not in the config
		if ovirt.Connection.CAFile == "" && ovirtEngineUrl.Scheme == "https" {
			fetchCafile(ovirt, ovirtEngineUrl.Hostname(), ovirtEngineUrl.Port())
		}
		rootCa, err := readCaCertPool(ovirt)
		if err != nil {
			return err
		}
		ovirt.client = http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: rootCa},
			},
		}
	}

	savedToken, err := ioutil.ReadFile(tokenStore)
	if err != nil {
		// ignore for now
	} else {
		json.Unmarshal(savedToken, &ovirt.token)
	}
	// get the token and persist if needed
	if ovirt.token.Value == "" || time.Now().After(ovirt.token.ExpirationTime) || !isTokenValid(ovirt) {
		ovirt.token, err = fetchToken(*ovirtEngineUrl, ovirt.Connection.Username, ovirt.Connection.Password, &ovirt.client)
		if err != nil {
			return err
		}
		persistToken(ovirt)
	}
	return nil
}

func persistToken(ovirt *Ovirt) {
	// store the fetched token
	j, _ := json.Marshal(ovirt.token)
	err := ioutil.WriteFile(tokenStore, j, 0600)
	if err != nil {
		// this err will be reported to stderr but won't bubble up.
		fmt.Fprintln(os.Stderr, err)
	}
}

// isTokenValid tries a simple GET / with the oauth token
// returns true for 200 ok, otherwise false
func isTokenValid(ovirt *Ovirt) bool {
	req, _ := getRequest(ovirt.Connection.Url, ovirt.token.Value)
	resp, err := ovirt.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode > 200 {
		return false
	}
	return true
}

// Attach will attach a disk to a VM on ovirt.
// nodeName is ovirt's vm name
// jsonParams is the volume info
// Response will include the device path according to the disk interface type
func (ovirt *Ovirt) Attach(params AttachRequest, nodeName string) (Response, error) {
	err := ovirt.Authenticate()
	// TODO validate params
	if err != nil {
		return FailedResponse, err
	}

	// convert params to json
	requestJson, err := json.Marshal(
		DiskAttachment{
			Bootable:    true,
			PassDiscard: true,
			Active:      true,
			Disk: Disk{
				Name: params.VolumeName,
				// TODO not in the spec, raise that
				ProvisionedSize: "1gb",
			},
		})

	if err != nil {
		return FailedResponse, err
	}

	// ovirt API call
	req, err := postWithJsonData(ovirt, "/vms/"+nodeName+"/diskattachments", requestJson)
	resp, err := ovirt.client.Do(req)

	if err != nil {
		return FailedResponse, err
	}
	defer resp.Body.Close()

	diskAttachment := DiskAttachment{}
	jsonResponse, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(jsonResponse, &diskAttachment)

	attachResponse := SuccessfulResponse
	shortDiskId := diskAttachment.Id[:16]
	switch diskAttachment.Interface {
	case "virtio":
		attachResponse.Device = "/dev/disk/by-id/virtio-" + shortDiskId
	case "virtio_iscsi":
		attachResponse.Device = "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_" + shortDiskId
	default:
		attachResponse.Message = "device type is unsupported"
	}

	return attachResponse, err
}

func readCaCertPool(ovirt *Ovirt) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(ovirt.Connection.CAFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool, nil
}

func fetchCafile(ovirt *Ovirt, hostname string, origPort string) error {
	port := "80"
	if origPort == "8443" {
		port = "8080"
	}
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/%s", hostname, port, caUrl))
	if err != nil {
		fmt.Println("Error while downloading CA", err)
		return err
	}
	defer resp.Body.Close()

	output, err := os.Create("ovirt.ca")
	if err != nil {
		return err
	}

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		return err
	}

	ovirt.Connection.CAFile = output.Name()
	return nil
}

// fetchToken will perform oauth password login to the engine to retrieve the token
// TODO write the token back to the config file so we don't need to perform login for every request
func fetchToken(ovirtEngineUrl url.URL, username string, password string, client *http.Client) (Token, error) {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s://%s/%s", ovirtEngineUrl.Scheme, ovirtEngineUrl.Host, tokenUrl),
		strings.NewReader(fmt.Sprintf(tokenPayload, username, password)),
	)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	//defer resp.Body.Close()

	if err != nil {
		return Token{}, err
	}

	tokenResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Token{}, err
	}

	t := Token{}
	err = json.Unmarshal(tokenResponse, &t)
	if err != nil {
		return Token{}, err
	}
	// TODO ovirt bug - ovirt always set the exp to Long.MAX_VALUE, nanosecs from epoch
	t.ExpirationTime = time.Unix(0, t.ExpireIn)
	return t, nil
}

func getRequest(endpoint string, t string) (*http.Request, error) {
	r, err := http.NewRequest("GET", endpoint, nil)
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+t)
	return r, err
}

func postWithJsonData(ovirt *Ovirt, endpoint string, json []byte) (*http.Request, error) {
	r, err := http.NewRequest(
		"POST",
		ovirt.Connection.Url+endpoint,
		strings.NewReader(string(json)),
	)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+ovirt.token.Value)
	return r, err
}
