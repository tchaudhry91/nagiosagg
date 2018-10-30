package svc

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/tchaudhry91/nagiosagg/parser"
)

// Instrumenting Middleware creates metrics from the underlying NagiosParserSvc

type instrumentingMiddleware struct {
	requests        metrics.Counter
	requestDuration metrics.Histogram
	numHosts        metrics.Histogram
	next            NagiosParserSvc
}

// GetParsedNagios instruments the underlying nagiosParseSvc endpoint
func (mw *instrumentingMiddleware) GetParsedNagios(ctx context.Context) (output map[string][]parser.NagiosStatus, err error) {
	defer func(begin time.Time) {
		lvs := []string{
			"method", "/nagios",
			"err", fmt.Sprint(err != nil),
		}
		mw.requests.With(lvs...).Add(1)
		mw.requestDuration.With(lvs...).Observe(time.Since(begin).Seconds())
		mw.numHosts.With(lvs...).Observe(float64(len(output)))
	}(time.Now())
	output, err = mw.next.GetParsedNagios(ctx)
	return output, err
}

func (mw *instrumentingMiddleware) RefreshNagiosData(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		lvs := []string{
			"method", "/refresh",
			"err", fmt.Sprint(err != nil),
		}
		mw.requests.With(lvs...).Add(1)
		mw.requestDuration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	err = mw.next.RefreshNagiosData(ctx)
	return err
}
