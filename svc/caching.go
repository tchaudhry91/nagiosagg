package svc

import (
	"context"

	cache "github.com/patrickmn/go-cache"
	"github.com/tchaudhry91/nagiosagg/parser"
)

// CachingMiddleware caches NagiosParserSvc requests and responses
type cachingMiddleware struct {
	cacher *cache.Cache
	next   NagiosParserSvc
}

// GetParsedNagios caches the values and proxies the request to the inner layer if not found
func (mw *cachingMiddleware) GetParsedNagios(ctx context.Context) (output map[string][]parser.NagiosStatus, err error) {
	var f interface{}
	var found bool
	if f, found = mw.cacher.Get("nagios"); !found {
		output, err = mw.next.GetParsedNagios(ctx)
		mw.cacher.Set("nagios", output, cache.DefaultExpiration)
		return output, err
	}
	output = f.(map[string][]parser.NagiosStatus)
	return output, nil
}

// RefreshNagiosData clears the cache and proxies the request to the inner layer
func (mw *cachingMiddleware) RefreshNagiosData(ctx context.Context) (err error) {
	defer func() {
		mw.cacher.Delete("nagios")
	}()
	err = mw.next.RefreshNagiosData(ctx)
	return err
}
