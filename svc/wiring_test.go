package svc

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var nagiosStatusDir = flag.String("nagios_status_dir", "statuses", "Nagios Status Directory")
var tempDBWire = filepath.Join(os.TempDir(), "wiring-test.db")
var srv *httptest.Server

func initService() {
	service, _ := NewNagiosParserSvc(*nagiosStatusDir, tempDBWire)
	router := MakeHTTPHandler(service)
	srv = httptest.NewServer(router)
}

func cleanUp() {
	os.Remove(tempDBWire)
}

func TestHTTPWiring(t *testing.T) {
	service, err := NewNagiosParserSvc(*nagiosStatusDir, tempDBWire)
	if err != nil {
		t.Errorf("Failed to build service")
		t.FailNow()
	}
	router := MakeHTTPHandler(service)
	if router == nil {
		t.Errorf("Failed to get handler")
		t.FailNow()
	}
	cleanUp()
}

func TestNagiosEndpoints(t *testing.T) {
	initService()
	for _, testcase := range []struct {
		method string
		url    string
		want   int
	}{
		{method: "GET", url: "/nagios", want: 500},
		{method: "GET", url: "/refresh", want: 200},
		{method: "GET", url: "/nagios", want: 200},
		{method: "GET", url: "/nagios2", want: 404},
	} {
		req, _ := http.NewRequest(testcase.method, srv.URL+testcase.url, nil)
		resp, _ := http.DefaultClient.Do(req)
		if want, have := testcase.want, resp.StatusCode; want != have {
			t.Errorf("%s %s: want %d, have %d", testcase.method, testcase.url, want, have)
		}
	}
	cleanUp()
}
