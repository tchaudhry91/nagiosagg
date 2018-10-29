package svc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kit/kit/endpoint"
	cache "github.com/patrickmn/go-cache"
)

type getParsedNagiosRequest struct{}

type getParsedNagiosResponse map[string][]NagiosStatusResponse

type refreshNagiosDataRequest struct{}

type refreshNagiosDataResponse struct {
	// Error doesn't JSON marshall, hence string
	Err string `json:"err,omitempty"`
}

// Endpoints is a struct containing all the endpoints for the NagiosParserService
type Endpoints struct {
	refreshNagiosData endpoint.Endpoint
	getParsedNagios   endpoint.Endpoint
}

// MakeServerEndpoints returns a struct with all the Endpoints for the NagiosParserService
func MakeServerEndpoints(svc NagiosParserSvc, cacher *cache.Cache) Endpoints {
	ee := Endpoints{}

	//gerParsedNagios Endpoint
	ee.getParsedNagios = MakeGetParsedNagiosEndpoint(svc)

	//refreshNagiosData Endpoint
	ee.refreshNagiosData = MakeRefreshNagiosDataEndpoint(svc)

	return ee
}

// NagiosStatusResponse is a filtered structure for Nagios data to be returned to the client
type NagiosStatusResponse struct {
	State            string    `json:"state,omitempty"`
	Output           string    `json:"output,omitempty"`
	Service          string    `json:"service,omitempty"`
	Attempts         string    `json:"attempts,omitempty"`
	LastCheck        time.Time `json:"last_check,omitempty"`
	NextCheck        time.Time `json:"next_check,omitempty"`
	LastStateChanged time.Time `json:"last_state_changed,omitempty"`
}

// MakeRefreshNagiosDataEndpoint returns an endpoint to refresh nagios data from new status files
func MakeRefreshNagiosDataEndpoint(svc NagiosParserSvc) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// req := request.(refreshNagiosDataRequest)
		// Skipped because empty request
		err := svc.RefreshNagiosData(ctx)
		resp := refreshNagiosDataResponse{}
		if err != nil {
			resp.Err = err.Error()
			return resp, err
		}
		return resp, nil
	}
}

// MakeGetParsedNagiosEndpoint returns an endpoint to get Parsed Nagios Data from multiple nagios instances
func MakeGetParsedNagiosEndpoint(svc NagiosParserSvc) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// req := request.(getParsedNagiosRequest)
		// Skipped because empty request
		resp, err := svc.GetParsedNagios(ctx)
		if err != nil {
			var issues getParsedNagiosResponse
			return issues, err
		}
		issues := getParsedNagiosResponse{}
		for host, problems := range resp {
			respIssues := []NagiosStatusResponse{}
			for _, problem := range problems {
				status := NagiosStatusResponse{}
				status.State = problem.State
				status.Service = problem.Service
				status.Output = problem.Values["plugin_output"]
				status.Attempts = fmt.Sprintf("%s/%s", problem.Values["current_attempt"], problem.Values["max_attempts"])
				loc, err := time.LoadLocation("UTC")
				if err != nil {
					return issues, nil
				}

				lastTS, err := strconv.ParseInt(problem.Values["last_check"], 10, 64)
				if err != nil {
					return issues, nil
				}
				status.LastCheck = time.Unix(lastTS, 0).In(loc)

				nextTS, err := strconv.ParseInt(problem.Values["next_check"], 10, 64)
				if err != nil {
					return issues, nil
				}
				status.NextCheck = time.Unix(nextTS, 0).In(loc)

				changeTS, err := strconv.ParseInt(problem.Values["last_state_change"], 10, 64)
				if err != nil {
					return issues, nil
				}
				status.LastStateChanged = time.Unix(changeTS, 0).In(loc)
				respIssues = append(respIssues, status)
			}
			issues[host] = respIssues
		}
		return issues, nil
	}
}
