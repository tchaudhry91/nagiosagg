package svc

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/tchaudhry91/nagiosagg/parser"
)

// LoggingMiddleware logs NagiosParserSvc requests and responses
type LoggingMiddleware struct {
	logger log.Logger
	next   NagiosParserSvc
}

// GetParsedNagios logs the values and proxies the request to the inner layer
func (mw LoggingMiddleware) GetParsedNagios(ctx context.Context) (output map[string][]parser.NagiosStatus, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "/nagios",
			"numhosts", len(output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	output, err = mw.next.GetParsedNagios(ctx)
	return output, err
}

// RefreshNagiosData logs the values and proxies the request to the inner layer
func (mw LoggingMiddleware) RefreshNagiosData(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "/refresh",
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	err = mw.next.RefreshNagiosData(ctx)
	return err
}

// NewLoggingMiddleware is a factory for the logging middleware to be used by the NagiosParserSvc
func NewLoggingMiddleware(logger log.Logger, next NagiosParserSvc) LoggingMiddleware {
	return LoggingMiddleware{logger, next}
}
