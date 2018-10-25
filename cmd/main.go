package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/tchaudhry91/nagiosagg/svc"
)

func main() {
	var (
		httpAddr        = flag.String("http.addr", ":8080", "HTTP listen address")
		nagiosStatusDir = flag.String("nagios_status_dir", "statuses", "Nagios Status Directory")
		localDB         = flag.String("local_db", filepath.Join(os.TempDir(), "nagios.db"), "Filepath to store nagios status data in")
	)
	flag.Parse()

	// Initialize Logger
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// Base Service
	service, err := svc.NewNagiosParserSvc(*nagiosStatusDir, *localDB)
	if err != nil {
		logger.Log("err", err.Error())
		panic("Failed to create service")
	}

	// Middlewares
	service = svc.LoggingMiddleware(logger)(service)

	// Initialize router
	r := svc.MakeHTTPHandler(service)

	http.ListenAndServe(*httpAddr, r)

}
