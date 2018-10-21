package svc

import (
	"context"
	"github.com/tchaudhry91/nagiosagg/parser"
)

// NagiosParserService is a service that returns aggregated data from various nagios sources
type NagiosParserService interface {
	GetNagiosData(ctx context.Context) ([]parser.NagiosStatus, error)
	RefreshNagiosData(ctx context.Context, statusDir string) error
}

type nagiosParserService struct {
	statusDir string
}
