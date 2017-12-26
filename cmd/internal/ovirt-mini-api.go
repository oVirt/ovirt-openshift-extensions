package internal

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/ovirt/ovirt-flexdriver/cmd/internal/model"
	"github.com/ovirt/ovirt-flexdriver/internal"
	"gopkg.in/gcfg.v1"
	"io"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/kubelet/client"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const caUrl = "ovirt-engine/services/pki-resource?resource=ca-certificate&format=X509-PEM-CA"
const tokenUrl = "/ovirt-engine/sso/oauth/token"
const tokenPayload = "grant_type=password&scope=ovirt-app-api&username=%s&password=%s"

var driverConfig = "ovirt-flexdriver.conf"

type Api struct {
	Connection Connection
	Client     http.Client
	Token      Token
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
	ExpireIn       int64  `json:"expires_in,string"`
	Type           string `json:"token_type"`
	ExpirationTime time.Time
}

// TODO move this to cmd, this is not ovirt mini api related
func InitDriver() (string, error) {
	value, exist := os.LookupEnv("OVIRT_FLEXDRIVER_CONF")
	if exist {
		driverConfig = value
	}
	var api Api
	if err := api.authenticate(); err != nil {
		return internal.FailedResponseJson, err
	}
	return internal.SuccesfullResonseJson, nil
}

func (api *Api) authenticate() error {
	err := gcfg.ReadFileInto(&api.Connection, driverConfig)
	if err != nil {
		return err
	}

	ovirtEngineUrl, err := url.Parse(api.Connection.Url)
	if err != nil {
		return err
	}

	if api.Connection.Insecure {
		api.Client = http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		// fetch ca if its not in the config
		if api.Connection.CAFile == "" && strings.HasPrefix(ovirtEngineUrl.Scheme, "https") {
			fetchCafile(api, ovirtEngineUrl.Host)
		}
		rootCa, err := readCaCertPool(api)
		if err != nil {
			return err
		}
		api.Client = http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: rootCa},
			},
		}
	}

	// get the token and store it
	if api.Token.Value == "" || time.Now().After(api.Token.ExpirationTime) {
		token, err := fetchToken(ovirtEngineUrl, api.Connection.Username, api.Connection.Password, &api.Client)
		if err != nil {
			return err
		}
		// authenticated successfully
		api.Token = token
		api.Token.ExpirationTime = time.Now().Add(time.Second * time.Duration(api.Token.ExpireIn))
	}
	return nil
}

// Attach will attach a disk to a VM on ovirt.
// nodeName is ovirt's vm name
// jsonParams is the volume info
// Response will include the device path according to the disk interface type
func (api *Api) Attach(params internal.AttachRequest, nodeName string) (internal.Response, error) {
	err := api.authenticate()
	// TODO validate params
	if err != nil {
		return internal.FailedResponse, err
	}

	// convert params to json
	requestJson, err := json.Marshal(
		model.DiskAttachment{
			Bootable:    true,
			PassDiscard: true,
			Active:      true,
			Disk: model.Disk{
				Name: params.VolumeName,
				// TODO not in the spec, raise that
				ProvisionedSize: "1gb",
			},
		})

	if err != nil {
		return internal.FailedResponse, err
	}

	// ovirt API call
	req, err := postWithJsonData(api, "/vms/"+nodeName+"/diskattachments", requestJson)
	resp, err := api.Client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return internal.FailedResponse, err
	}

	diskAttachment := model.DiskAttachment{}
	jsonResponse, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(jsonResponse, &diskAttachment)

	attachResponse := internal.SuccessfulResponse
	shortDiskId := diskAttachment.Id[:16]
	switch diskAttachment.Interface {
	case "virtio":
		attachResponse.Device = "/dev/disk/by-id/virtio-" + shortDiskId
	case "virtio_iscsi":
		attachResponse.Device = "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_" + shortDiskId
	default:
		fmt.Errorf("device type is unsupported")
	}

	return attachResponse, err
}

func (api *Api) GetVolumeByName(vmName, volName string) (string, error) {
	err := api.authenticate()
	if err != nil {
		return "", err
	}

	r, err := getRequest(fmt.Sprintf(api.Connection.Url+"/disks?search=name=/%s/diskattachments?search", volName))
	response, err := api.Client.Do(r)

	if err != nil {
		return "", err
	}

	if response.StatusCode == 200 {
		dec := json.NewDecoder(response.Body)
		var m map[string]interface{}
		for {
			if err := dec.Decode(&m); err == io.EOF {
				break
			} else if err != nil {
				return "", err
			}
		}
		for k, v := range m {
			if k == ""
		}
	}
}

func readCaCertPool(api *Api) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(api.Connection.CAFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool, nil
}

func fetchCafile(api *Api, host string) error {
	resp, err := http.Get("http://" + host + caUrl)
	if err != nil {
		fmt.Println("Error while downloading CA", err)
		return err
	}
	defer resp.Body.Close()

	output, err := os.Create("ovirt.ca")
	if err != nil {
		fmt.Errorf("error %s", err)
		return err
	}

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		fmt.Errorf("error %s", err)
		return err
	}

	api.Connection.CAFile = output.Name()
	return nil
}

// fetchToken will perform oauth password login to the engine will retrieve the token respones
// TODO write the token back to the config file so we don't need to perform login for every request
func fetchToken(ovirtEngineUrl *url.URL, username string, password string, client *http.Client) (token Token, err error) {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/%s", ovirtEngineUrl, tokenUrl),
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
	return t, nil
}

func getRequest(endpoint string) (*http.Request, error) {
	r, err := http.NewRequest("GET", endpoint, nil)
	r.Header.Set("Accept", "application/json")
	return r, err
}

func postWithJsonData(api *Api, endpoint string, json []byte) (*http.Request, error) {
	r, err := http.NewRequest(
		"POST",
		api.Connection.Url+endpoint,
		strings.NewReader(string(json)),
	)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	return r, err
}
