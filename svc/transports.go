package svc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	cache "github.com/patrickmn/go-cache"
)

var (
	//ErrJSONUnMarshall indicates a bad request where json unmarshalling failed
	ErrJSONUnMarshall = errors.New("failed to parse json")
)

// MakeHTTPHandler returns an http handler for the endpoints
func MakeHTTPHandler(svc NagiosParserSvc, cacher *cache.Cache) http.Handler {
	r := mux.NewRouter()
	ee := MakeServerEndpoints(svc, cacher)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}
	getParsedNagiosHandler := httptransport.NewServer(
		ee.getParsedNagios,
		decodeGetParsedNagiosRequest,
		encodeGetParsedNagiosResponse,
		options...,
	)
	r.Methods("GET").Path("/nagios").Handler(getParsedNagiosHandler)

	refreshNagiosDataHandler := httptransport.NewServer(
		ee.refreshNagiosData,
		decodeRefreshNagiosDataRequest,
		encodeRefreshNagiosDataResponse,
		options...,
	)
	r.Methods("GET").Path("/refresh").Handler(refreshNagiosDataHandler)
	return r
}
func decodeRefreshNagiosDataRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// We return a blank request holder because no data must be taken in yet
	return refreshNagiosDataRequest{}, nil
}

func encodeRefreshNagiosDataResponse(_ context.Context, w http.ResponseWriter, resp interface{}) error {
	return json.NewEncoder(w).Encode(resp)
}

func decodeGetParsedNagiosRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// We return a blank request holder because no data must be taken in yet
	return getParsedNagiosRequest{}, nil
}

func encodeGetParsedNagiosResponse(_ context.Context, w http.ResponseWriter, resp interface{}) error {
	return json.NewEncoder(w).Encode(resp)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrJSONUnMarshall:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
