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
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log/syslog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const caUrl = "ovirt-engine/services/pki-resource?resource=ca-certificate&format=X509-PEM-CA"
const tokenUrl = "/ovirt-engine/sso/oauth/token"
const tokenPayload = "grant_type=password&scope=ovirt-app-api&username=%s&password=%s"

var tokenStore = "/tmp/ovirt-flexdriver.token"
var log, logError = syslog.New(syslog.LOG_INFO, "ovirt-api")

type Ovirt struct {
	Connection Connection
	client     http.Client
	token      Token
}

type Connection struct {
	Url      string
	Username string
	Password string
	Insecure bool
	CAFile   string
}

type Token struct {
	Value          string `json:"access_token"`
	ExpireIn       int64  `json:"exp,string"`
	Type           string `json:"token_type"`
	ExpirationTime time.Time
}

// newDriver creates a new ovirt driver instance from a config reader, to make it
// easy to pass various config items, either file, string, reading from remote etc.
// the underlying config format supports properties files (like java)
func NewOvirt(configReader io.Reader) (OvirtApi, error) {
	viper.SetConfigType("props")
	o := Ovirt{}
	if err := viper.ReadConfig(configReader); err != nil {
		return nil, err
	}
	o.Connection.Url = viper.GetString("url")
	o.Connection.Username = viper.GetString("username")
	o.Connection.Password = viper.GetString("password")
	o.Connection.Insecure = viper.GetBool("insecure")
	o.Connection.CAFile = viper.GetString("cafile")
	return &o, nil
}

func (ovirt *Ovirt) GetConnectionDetails() Connection {
	return ovirt.Connection
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
		// ignore, probably first invocation or a re-authentication
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
		logErrorf(fmt.Sprintf("error persisting token %s", err))
	}
}

// isTokenValid tries a simple GET / with the oauth token
// returns true for 200 ok, otherwise false
func isTokenValid(ovirt *Ovirt) bool {
	resp, err := ovirt.clientDo(http.MethodGet, ovirt.Connection.Url, strings.NewReader(""))

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode > 200 {
		return false
	}
	return true
}

func (ovirt *Ovirt) GetDiskByName(diskName string) (DiskResult, error) {
	var diskResult DiskResult
	r, err := ovirt.Get(fmt.Sprintf("disks?search=name=%s", diskName))
	if err != nil {
		return diskResult, err
	}
	err = json.Unmarshal(r, &diskResult)
	return diskResult, err
}

func (ovirt *Ovirt) CreateUnattachedDisk(diskName string, storageDomainName string, sizeIbBytes int64, readOnly bool, thinProvisioning bool) (Disk, error) {
	format, sparse, err := ovirt.DefaultDiskParamsBy(storageDomainName, thinProvisioning)
	if err != nil {
		return Disk{}, err
	}
	disk := Disk{
		Name:            diskName,
		ProvisionedSize: uint64(sizeIbBytes),
		Format:          format,
		StorageDomains:  StorageDomains{[]StorageDomain{{Name: storageDomainName}}},
		Sparse:          sparse,
	}

	post, err := ovirt.Post("disks", disk)
	if err != nil {
		return disk, err
	}
	result := Disk{}
	err = json.Unmarshal([]byte(post), &result)
	return result, err
}

// this logic is aligned with oVirt logic for determining disk format and spareness
// the combination are determined by the type of the storage domain.
func (ovirt *Ovirt) DefaultDiskParamsBy(storageDomainName string, thinProvisioned bool) (DiskFormat, Sparse, error){

	if !thinProvisioned {
		// default no matter what the disk is - raw disk, no sparseness
		return "raw", false, nil
	}

	domain, e := ovirt.GetStorageDomainBy(storageDomainName)
	if e != nil {
		return "", false, e
	}

	// thin provisioned
	// block (iscsi/fc)    - cow + sparse
	// file  (nfs/gluster) - raw + sparse
	if domain.Storage.Type == "iscsi" || domain.Storage.Type == "fc" {
		return "cow", true, nil
	}
	return "raw", true, nil
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

	post, err := ovirt.Post("vms/"+vmId+"/diskattachments", a)
	if err != nil {
		return a, err
	}
	r := DiskAttachment{}
	err = json.Unmarshal([]byte(post), &r)
	return r, err
}

func (ovirt *Ovirt) Get(path string) ([]byte, error) {
	resp, err := ovirt.clientDo(http.MethodGet, path, strings.NewReader(""))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 200 {
		return nil, translateError(*resp)
	}

	b, err := ioutil.ReadAll(resp.Body)
	return b, err
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

func (ovirt *Ovirt) Post(path string, data interface{}) (string, error) {
	d, err := json.Marshal(data)
	if err != nil {
		// failed json conversion
		return "", err
	}
	resp, err := ovirt.clientDo(http.MethodPost, path, strings.NewReader(string(d)))

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 300 {
		return "", errors.New(resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	return string(b), err
}

func (ovirt *Ovirt) Delete(path string) ([]byte, error) {
	resp, err := ovirt.clientDo(http.MethodDelete, path, strings.NewReader(""))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 200 {
		return nil, errors.New(resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	return b, err
}

//TODO implement with invocation of the new GetVMs
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

func (ovirt *Ovirt) GetVMById(id string) (VM, error) {
	s, err := ovirt.Get("vms/" + id)
	vm := VM{}
	if err != nil {
		return VM{}, err
	}
	err = json.Unmarshal([]byte(s), &vm)
	return vm, err
}

func (ovirt *Ovirt) GetVMs(searchQuery string) ([]VM, error) {
	s, err := ovirt.Get(searchQuery)
	vmResult := VMResult{}
	if err != nil {
		return vmResult.Vms, err
	}
	err = json.Unmarshal([]byte(s), &vmResult)
	return vmResult.Vms, err
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

	if err != nil {
		return Token{}, err
	}

	defer resp.Body.Close()

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

func (ovirt *Ovirt) clientDo(method string, url string, payload io.Reader) (*http.Response, error) {
	url = fmt.Sprintf("%s/%s", ovirt.Connection.Url, url)
	logInfof("calling ovirt api url: %s", url)
	r, _ := http.NewRequest(method, url, payload)
	r.Header.Set("Accept", "application/json")
	r.Header.Add("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+ovirt.token.Value)

	resp, err := ovirt.client.Do(r)

	if err != nil {
		logErrorf("failed to call ovirt api: %s", err)
		return resp, err
	}

	if resp.StatusCode >= 300 {
		logInfof("failed to call ovirt api with response: %s", resp.Body)
		if resp.StatusCode == 401 {
			// invalid token, probably expired due to inactivity or
			// ovirt-engine has restarted. ovirt-engine doesn't support
			// fully persistent oauth tokens
			logInfof("ovirt api rejected the token, re-authenticating...")
			err := os.Remove(tokenStore)
			if err != nil {
				logInfof("failed to remove the old token file %s", err)
			}
			ovirt.token.Value = ""
			ovirt.Authenticate()
		}
	}

	return resp, err
}

// GetStorageDomainBy returns a storage domain type by name
func (ovirt *Ovirt) GetStorageDomainBy(name string) (StorageDomain, error){

	s, err := ovirt.Get("storagedomains?search=name=" + name)
	domains := StorageDomains{}
	if err != nil {
		return StorageDomain{}, err
	}
	err = json.Unmarshal([]byte(s), &domains)
	if err != nil {
		return StorageDomain{}, err
	}
	if len(domains.Domains) > 0 {
		return domains.Domains[0], nil
	}

	return StorageDomain{}, ErrNotExist

}

func logInfof(format string, message ...interface{}) {
	if log != nil && logError == nil {
		log.Info(fmt.Sprintf(format, message...))
	}
}

func logErrorf(format string, message ...interface{}) {
	if log != nil && logError == nil {
		log.Err(fmt.Sprintf(format, message...))
	}
}
