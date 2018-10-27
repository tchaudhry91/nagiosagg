package svc

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	cache "github.com/patrickmn/go-cache"
)

var nagiosStatusDir = flag.String("nagios_status_dir", "statuses", "Nagios Status Directory")
var tempDBWire = filepath.Join(os.TempDir(), "wiring-test.db")

func initService() *httptest.Server {
	logger := log.NewNopLogger()
	cacher := cache.New(3*time.Minute, 3*time.Minute)
	service, _ := NewNagiosParserSvc(*nagiosStatusDir, tempDBWire)
	service = LoggingMiddleware(logger)(service)
	router := MakeHTTPHandler(service, cacher)
	return httptest.NewServer(router)
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
	service = LoggingMiddleware(log.NewNopLogger())(service)
	cacher := cache.New(3*time.Minute, 3*time.Minute)
	router := MakeHTTPHandler(service, cacher)
	if router == nil {
		t.Errorf("Failed to get handler")
		t.FailNow()
	}
	cleanUp()
}

func TestNagiosEndpoints(t *testing.T) {
	srv := initService()
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

func TestEndpointTiming(t *testing.T) {
	srv := initService()
	for _, testcase := range []struct {
		name   string
		method string
		url    string
		want   int
	}{
		{name: "BenchRefresh", method: "GET", url: "/refresh", want: 200},
		{name: "BenchFirstData", method: "GET", url: "/nagios", want: 200},
		{name: "BenchCachedData", method: "GET", url: "/nagios", want: 200},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			req, _ := http.NewRequest(testcase.method, srv.URL+testcase.url, nil)
			resp, _ := http.DefaultClient.Do(req)
			if want, have := testcase.want, resp.StatusCode; want != have {
				t.Errorf("%s %s: want %d, have %d", testcase.method, testcase.url, want, have)
			}
		})
	}
	cleanUp()
}
