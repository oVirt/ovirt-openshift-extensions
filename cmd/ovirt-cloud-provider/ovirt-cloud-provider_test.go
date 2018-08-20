package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var vmsJsonData string

const vm1Id = "0013e1a7-c837-4b7f-8420-6deca9486415"
const vm2Id = "e47839a3-149b-4405-ba69-3fb20eaa2fed"
const vm1NodeName = "ovirtNode_1"

type HttpHandler struct{}

func TestMain(m *testing.M) {
	if vmsJsonData == "" {
		parse, err := ioutil.ReadFile("./vms.json")
		if err != nil {
			panic(err)
		}
		vmsJsonData = string(parse)
	}
	m.Run()
}

func TestNewProvider(t *testing.T) {
	p, err := NewOvirtProvider(ProviderConfig{})
	if err != nil || p == nil {
		t.Fatal(err)
	}
}

// TestGetInstanceId test the id returned from the api call
func TestGetInstanceId(t *testing.T) {
	// mock the api call to return a json of vms
	provider, err := getProvider()

	id, err := provider.InstanceID(nil, vm1NodeName)
	if err != nil {
		t.Fatal(err)
	}

	if id == "" {
		t.Fatal(err)
	}

	if id != vm1Id {
		t.Fatalf("expected id %s is no equal to %s", vm1Id, id)
	}
}

func TestProvider_CurrentNodeName(t *testing.T) {
	p, _ := getProvider()
	name, err := p.CurrentNodeName(nil, vm1NodeName)
	if err != nil || name == "" {
		t.Fatalf("node name %s was not found. Error was %s", vm1NodeName, err)
	}
}

func TestProvider_InstanceExistsByProviderId(t *testing.T) {
	p, _ := getProvider()
	exists, _ := p.InstanceExistsByProviderID(nil, vm1Id)
	if !exists {
		t.Fatalf("the instance %s should exist with status which is other than down", vm1Id)
	}
}

func TestProvider_InstanceByProviderIdNotExists(t *testing.T) {
	p, _ := getProvider()
	exists, _ := p.InstanceExistsByProviderID(nil, vm2Id)
	if exists {
		t.Fatalf("the instance %s should not exist with status down", vm2Id)
	}
}

func TestProvider_NodeAddressesByProviderID(t *testing.T) {
	p, _ := getProvider()
	p.NodeAddresses(nil, vm1NodeName)
}

func getProvider() (*CloudProvider, error) {
	httpServer := mockGetVms()
	//TODO How to defer this using some Before-After hooks of tests in go? I can't defer in place otherwise tests will not
	// be able to use it (it closes before the test actually uses it)
	//defer httpServer.Close()
	c := ProviderConfig{}
	c.Connection.Url = httpServer.URL
	provider, err := NewOvirtProvider(c)
	return provider, err
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	i := string(vmsJsonData)
	io.WriteString(w, i)
}

func mockGetVms() *httptest.Server {
	server := httptest.NewServer(&HttpHandler{})
	return server
}
