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
	"golang.org/x/time/rate"
)

var nagiosStatusDir = flag.String("nagios_status_dir", "statuses", "Nagios Status Directory")
var tempDBWire = filepath.Join(os.TempDir(), "wiring-test.db")

const limitInterval time.Duration = 20

func initService() *httptest.Server {
	logger := log.NewNopLogger()
	cacher := cache.New(3*time.Minute, 3*time.Minute)
	limit := rate.Every(time.Second * limitInterval)
	limiter := rate.NewLimiter(limit, 1)

	service, _ := NewNagiosParserSvc(*nagiosStatusDir, tempDBWire)
	service = LoggingMiddleware(logger)(service)
	service = CachingMiddleware(cacher)(service)
	router := MakeHTTPHandler(service, cacher, limiter)
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
	limiter := rate.NewLimiter(rate.Every(time.Minute), 1)
	router := MakeHTTPHandler(service, cacher, limiter)
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

func TestRateLimiter(t *testing.T) {
	srv := initService()
	for _, testcase := range []struct {
		name   string
		method string
		url    string
		want   int
		sleep  time.Duration
	}{
		{name: "FirstRefresh", method: "GET", url: "/refresh", want: 200, sleep: 0},
		{name: "SecondRefresh", method: "GET", url: "/refresh", want: 429, sleep: 0},
		{name: "ThirdRefresh", method: "GET", url: "/refresh", want: 200, sleep: limitInterval * time.Second},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			time.Sleep(testcase.sleep)
			req, _ := http.NewRequest(testcase.method, srv.URL+testcase.url, nil)
			resp, _ := http.DefaultClient.Do(req)
			if want, have := testcase.want, resp.StatusCode; want != have {
				t.Errorf("%s %s: want %d, have %d", testcase.method, testcase.url, want, have)
			}
		})
	}
	cleanUp()
}
