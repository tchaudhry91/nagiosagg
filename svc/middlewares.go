package svc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	cache "github.com/patrickmn/go-cache"
)

// Middleware is a service middleware builder
type Middleware func(NagiosParserSvc) NagiosParserSvc

// LoggingMiddleware produces a logging middleware builder. This is a service middleware
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next NagiosParserSvc) NagiosParserSvc {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

// CachingMiddleware produces a caching wrapper for an endpoint
func cachingMiddleware(cacher *cache.Cache) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			var f interface{}
			var found bool
			if f, found = cacher.Get("nagios"); !found {
				ret, err := next(ctx, request)
				ret = ret.(getParsedNagiosResponse)
				cacher.Set("nagios", ret, cache.DefaultExpiration)
				return ret, err
			}
			return f.(getParsedNagiosResponse), nil
		}
	}
}
