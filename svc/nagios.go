package svc

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/tchaudhry91/nagiosagg/parser"
)

// NagiosParserSvc is a service that returns aggregated data from various nagios sources
type NagiosParserSvc interface {
	GetParsedNagios(ctx context.Context) (map[string][]parser.NagiosStatus, error)
	//	RefreshNagiosData(ctx context.Context, statusDir string) error
}

type nagiosParserSvc struct {
	statusDir string
}

//GetParsedNagiosData returns a parsed map of hostname to issues from various nagios status files
func (svc nagiosParserSvc) GetParsedNagios(ctx context.Context) (map[string][]parser.NagiosStatus, error) {
	result := make(map[string][]parser.NagiosStatus)
	files, err := filepath.Glob(svc.statusDir + "/*.dat")
	if err != nil {
		return result, err
	}
	gatherers := len(files)
	var wg sync.WaitGroup
	resultChan := make(chan map[string][]parser.NagiosStatus, gatherers)
	errChan := make(chan error, gatherers)

	for _, f := range files {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			resultsLocal, errLocal := parser.ParseStatusFromFile(filename)
			if errLocal != nil {
				errChan <- errLocal
			}
			resultChan <- resultsLocal
		}(f)
	}
	wg.Wait()
	close(resultChan)
	close(errChan)
	if len(errChan) > 0 {
		return result, fmt.Errorf("Failed to parse nagios data: %v ", <-errChan)
	}
	for resultChunk := range resultChan {
		for hostname, values := range resultChunk {
			result[hostname] = values
		}
	}
	return result, nil
}
