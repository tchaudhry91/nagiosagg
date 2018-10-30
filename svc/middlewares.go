package svc

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
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

// CachingMiddleware produces a caching middleware builder. This is a service middleware
func CachingMiddleware(cacher *cache.Cache) Middleware {
	return func(next NagiosParserSvc) NagiosParserSvc {
		return &cachingMiddleware{
			next:   next,
			cacher: cacher,
		}
	}
}

// InstrumentingMiddleware produces an instrumenting middleware builder. This is a service middleware
func InstrumentingMiddleware(requests metrics.Counter, requestDuration metrics.Histogram, numHosts metrics.Histogram) Middleware {
	return func(next NagiosParserSvc) NagiosParserSvc {
		return &instrumentingMiddleware{
			next:            next,
			requests:        requests,
			requestDuration: requestDuration,
			numHosts:        numHosts,
		}
	}
}
