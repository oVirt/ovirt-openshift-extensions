package internal

import (
	"net/http"
	"net/http/httptest"
)

// MockOvirt has a multiplexer inside to register multi request handler during
// a test which performs few request
type MockOvirt struct {
	*Ovirt
	ServeMux *http.ServeMux
}

func NewMockOvirt() MockOvirt {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	ovirt := Ovirt{
		Connection: Connection{
			Url:      ts.URL,
			Insecure: true,
		},
		client: *http.DefaultClient,
	}
	return MockOvirt{
		Ovirt:    &ovirt,
		ServeMux: mux,
	}
}

func (mockOvirt *MockOvirt) Handle(path string, handler http.HandlerFunc){
	mockOvirt.ServeMux.HandleFunc(path, handler)
}

func CreateMockOvirtClient(handler http.HandlerFunc) Ovirt {
	ts := httptest.NewServer(handler)
	return Ovirt{
		Connection: Connection{
			Url:      ts.URL,
			Insecure: true,
		},
		client: *http.DefaultClient,
	}
}
