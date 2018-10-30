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
	kitprom "github.com/go-kit/kit/metrics/prometheus"
	cache "github.com/patrickmn/go-cache"
	stdprom "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

var nagiosStatusDir = flag.String("nagios_status_dir", "statuses", "Nagios Status Directory")
var tempDBWire = filepath.Join(os.TempDir(), "wiring-test.db")

const limitInterval time.Duration = 20

var requests *kitprom.Counter
var requestDuration *kitprom.Summary
var numHosts *kitprom.Summary

func init() {
	fieldKeys := []string{"method", "err"}
	requests = kitprom.NewCounterFrom(
		stdprom.CounterOpts{
			Namespace: "nagios_svc",
			Name:      "requests_count",
			Help:      "Total Endpoints Requested",
		},
		fieldKeys,
	)
	requestDuration = kitprom.NewSummaryFrom(
		stdprom.SummaryOpts{
			Namespace: "nagios_svc",
			Name:      "request_duration",
			Help:      "Time taken per request",
		},
		fieldKeys,
	)
	numHosts = kitprom.NewSummaryFrom(
		stdprom.SummaryOpts{
			Namespace: "nagios_svc",
			Name:      "num_hosts",
			Help:      "Number of hosts found with issues",
		},
		fieldKeys,
	)
}

func initService() *httptest.Server {
	cleanUp()
	// Middleware inits
	logger := log.NewNopLogger()
	cacher := cache.New(3*time.Minute, 3*time.Minute)
	limit := rate.Every(time.Second * limitInterval)
	limiter := rate.NewLimiter(limit, 1)

	// Service inits
	service, _ := NewNagiosParserSvc(*nagiosStatusDir, tempDBWire)
	service = CachingMiddleware(cacher)(service)
	service = InstrumentingMiddleware(requests, requestDuration, numHosts)(service)
	service = LoggingMiddleware(logger)(service)
	router := MakeHTTPHandler(service, cacher, limiter)

	return httptest.NewServer(router)
}

func cleanUp() {
	os.Remove(tempDBWire)
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
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
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

func BenchmarkGetNagiosDataRequests(b *testing.B) {
	srv := initService()
	// Populate Data
	req, _ := http.NewRequest("GET", srv.URL+"/refresh", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		b.Errorf("Refresh failed")
	}
	b.ResetTimer()
	var errCount int
	for n := 0; n < b.N; n++ {
		req, _ := http.NewRequest("GET", srv.URL+"/nagios", nil)
		resp, _ := http.DefaultClient.Do(req)
		if resp.StatusCode != http.StatusOK {
			errCount++
		}
	}
	b.Logf("Errors:%d", errCount)
}
