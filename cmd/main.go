package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/kit/log"
	cache "github.com/patrickmn/go-cache"
	"github.com/tchaudhry91/nagiosagg/svc"
	"golang.org/x/time/rate"
)

func main() {
	var (
		httpAddr        = flag.String("http.addr", ":8080", "HTTP listen address")
		nagiosStatusDir = flag.String("nagios_status_dir", "statuses", "Nagios Status Directory")
		localDB         = flag.String("local_db", filepath.Join(os.TempDir(), "nagios.db"), "Filepath to store nagios status data in")
		refreshTime     = flag.Int64("cache_expiration", 180, "Seconds to keep results cached")
		rateLimiter     = flag.Int64("refresh_interval", 60, "Minimum seconds between processing refresh requests")
	)
	flag.Parse()
	// Initialize Logger
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// Initialize in-mem cacher
	cacher := cache.New(time.Duration(*refreshTime)*time.Second, time.Duration(*refreshTime)*time.Second)

	// Initialize refresh rate limiter
	limiter := rate.NewLimiter(rate.Every(time.Duration(*rateLimiter)*time.Second), 1)

	// Base Service
	service, err := svc.NewNagiosParserSvc(*nagiosStatusDir, *localDB)
	if err != nil {
		logger.Log("err", err.Error())
		panic("Failed to create service")
	}

	// Middlewares

	service = svc.CachingMiddleware(cacher)(service)
	service = svc.LoggingMiddleware(logger)(service)

	// Initialize router
	r := svc.MakeHTTPHandler(service, cacher, limiter)

	http.ListenAndServe(*httpAddr, r)

}
