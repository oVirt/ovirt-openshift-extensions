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
	"errors"
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
const tokenStore = "/tmp/ovirt-flexdriver.token"

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
// vmName is ovirt's vm name
// jsonParams is the volume info
// Response will include the device path according to the disk interface type
func (ovirt *Ovirt) Attach(params AttachRequest, vmName string) (Response, error) {
	// TODO validate params
	// convert params to json
	requestJson, err := json.Marshal(
		DiskAttachment{
			Bootable:    true,
			PassDiscard: true,
			Active:      true,
			Disk: Disk{
				Name: params.VolumeName,
			},
		})

	if err != nil {
		return FailedResponse, err
	}

	// ovirt API call
	req, err := postWithJsonData(ovirt, "/vms/"+vmName+"/diskattachments", requestJson)
	resp, err := ovirt.client.Do(req)

	if err != nil {
		return FailedResponse, err
	}
	defer resp.Body.Close()

	diskAttachment := DiskAttachment{}
	jsonResponse, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(jsonResponse, &diskAttachment)

	attachResponse := SuccessfulResponse

	return attachResponse, err
}
func (ovirt Ovirt) GetDiskByName(diskName string) (DiskResult, error) {
	var diskResult DiskResult
	r, err := ovirt.Get(fmt.Sprintf("disks?search=name=%s", diskName))
	if err != nil {
		return diskResult, err
	}
	err = json.Unmarshal(r, &diskResult)
	return diskResult, err
}

func (ovirt *Ovirt) CreateUnattachedDisk(diskName string, storageDomainName string, sizeIbBytes int64, readOnly bool, format string) (Disk, error) {
	disk := Disk{
		Name:            diskName,
		ProvisionedSize: uint64(sizeIbBytes),
		Format:          "raw",
		StorageDomains:  StorageDomains{[]StorageDomain{{Name: storageDomainName}}},
	}

	post, err := ovirt.Post("/disks", disk)
	if err != nil {
		return disk, err
	}
	result := Disk{}
	err = json.Unmarshal([]byte(post), &result)
	return result, err
}

func (ovirt *Ovirt) CreateDisk(
	diskName string,
	storageDomainName string,
	readOnly bool,
	vmId string,
	diskId string,
	diskInterface string) (DiskAttachment, error) {

	a := DiskAttachment{
		Active: true,
		Disk: Disk{
			Name:   diskName,
			Format: "raw",
			StorageDomains: StorageDomains{
				[]StorageDomain{{Name: storageDomainName}},
			},
		},
		ReadOnly: readOnly,
	}
	if diskInterface != "" {
		a.Interface = diskInterface
	}
	if diskInterface == "" {
		a.Interface = "virtio_scsi"
	}
	if diskId != "" {
		a.Disk.Id = diskId
	}

	post, err := ovirt.Post("/vms/"+vmId+"/diskattachments", a)
	if err != nil {
		return a, err
	}
	r := DiskAttachment{}
	err = json.Unmarshal([]byte(post), &r)
	return r, err
}

func (ovirt Ovirt) Get(path string) ([]byte, error) {
	request, err := getRequest(fmt.Sprintf("%s/%s", ovirt.Connection.Url, path), ovirt.token.Value)
	if err != nil {
		return nil, err
	}
	response, err := ovirt.client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode > 200 {
		return nil, translateError(*response)
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	return bytes, err
}

type NotFound struct {
	response http.Response
}

func (n NotFound) Error() string {
	return fmt.Sprintf("No resource at " + n.response.Request.URL.Path)
}

func translateError(response http.Response) error {
	switch response.StatusCode {
	case 404:
		return NotFound{response: response}
	}
	return errors.New(response.Status)
}

func (ovirt Ovirt) Post(path string, data interface{}) (string, error) {
	d, err := json.Marshal(data)
	if err != nil {
		// failed json conversion
		return "", err
	}
	fmt.Println(string(d))
	request, err := postWithJsonData(&ovirt, path, d)
	if err != nil {
		return "", err
	}
	response, err := ovirt.client.Do(request)
	if err != nil {
		return "", err
	}
	if response.StatusCode > 300 {
		return "", errors.New(response.Status)
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	return string(bytes), err
}

func (ovirt Ovirt) Delete(path string) ([]byte, error) {
	request, err := deleteRequest(fmt.Sprintf("%s/%s", ovirt.Connection.Url, path), ovirt.token.Value)
	if err != nil {
		return nil, err
	}
	response, err := ovirt.client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode > 200 {
		return nil, errors.New(response.Status)
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	return bytes, err
}

func (ovirt *Ovirt) GetVM(name string) (VM, error) {
	s, err := ovirt.Get("vms?search=name=" + name)
	vmResult := VMResult{}
	if err != nil {
		return VM{}, err
	}
	err = json.Unmarshal([]byte(s), &vmResult)
	var vm VM
	if len(vmResult.Vms) > 0 {
		vm = vmResult.Vms[0]
	}
	return vm, err

}
func (ovirt *Ovirt) GetDiskAttachment(vmId, diskId string) (DiskAttachment, error) {
	s, err := ovirt.Get("vms/" + vmId + "/diskattachments/" + diskId)
	d := DiskAttachment{}
	if err != nil {
		return d, err
	}
	err = json.Unmarshal([]byte(s), &d)
	return d, err
}

func (ovirt *Ovirt) GetDiskAttachments(vmId string) ([]DiskAttachment, error) {
	s, err := ovirt.Get("vms/" + vmId + "/diskattachments/")
	result := DiskAttachmentResult{}
	if err != nil {
		return result.DiskAttachments, err
	}
	err = json.Unmarshal([]byte(s), &result)
	return result.DiskAttachments, err
}

func (ovirt *Ovirt) DetachDiskFromVM(vmId string, diskId string) error {
	_, err := ovirt.Delete("vms/" + vmId + "/diskattachments/" + diskId)
	return err
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

	if resp.StatusCode != 200 {
		return Token{}, fmt.Errorf("fail to login and fetching token %s", resp.Status)
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

func getRequest(endpoint string, token string) (*http.Request, error) {
	r, err := http.NewRequest("GET", endpoint, nil)
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+token)
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

func deleteRequest(endpoint string, token string) (*http.Request, error) {
	r, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+token)
	return r, err
}
