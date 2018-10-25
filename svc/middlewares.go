package svc

import "github.com/go-kit/kit/log"

// Middleware is a service middleware builder
type Middleware func(NagiosParserSvc) NagiosParserSvc

// LoggingMiddleware produces a logging middleware builder
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next NagiosParserSvc) NagiosParserSvc {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}
