# Nagios Aggregator Service
[![Go Report Card](https://goreportcard.com/badge/github.com/tchaudhry91/nagiosagg)](https://goreportcard.com/report/github.com/tchaudhry91/nagiosagg)

This is a simple service that maintains an aggregation of nagios alerts from different nagios instances.
It works by parsing the nagios status.dat files and returns a JSON response containing a map with lists of issues mapped per hostname.


**Usage**:
```
Usage of ./nagios:
  -cache_expiration int
        Seconds to keep results cached (default 180)
  -http.addr string
        HTTP listen address (default ":8080")
  -local_db string
        Filepath to store nagios status data in (default "/tmp/nagios.db")
  -nagios_status_dir string
        Directory containing .dat files from nagios (default "statuses")
  -refresh_interval int
        Minimum seconds between processing refresh requests (default 60)
```

**Endpoints**:

```
GET /refresh
```
The `/refresh` endpoint parsed the status.dat data and updated the local_db. Since this can be an intensive operation, it can be rate limited by `-refresh_interval`

```
GET /nagios
```
The `/nagios` endpoint returns a JSON with entries as follows:
```
{
    "hostname1": [
        {
            "state": "WARNING",
            "output": "plugin output goes here",
            "service": "xyz service",
            "attempts": "4/4",
            "last_check": "2018-09-26T06:15:41Z",
            "next_check": "2018-09-26T06:30:41Z",
            "last_state_changed": "2018-08-29T06:41:14Z"
        },
        {
            "state": "CRITICAL",
            "output": "plugin output goes here",
            "service": "abc service",
            "attempts": "4/4",
            "last_check": "2018-09-26T06:15:41Z",
            "next_check": "2018-09-26T06:30:41Z",
            "last_state_changed": "2018-08-29T06:41:14Z"
        }
    ],
    "hostname2": [
        ...
    ]
}
```

The `/metrics` endpoint returns prometheus format metrics for the service

Tanmay Chaudhry (tanmay.chaudhry@gmail.com)
