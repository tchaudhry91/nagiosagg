package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/tchaudhry91/nagiosagg/parser"
)

// NagiosParserSvc is a service that returns aggregated data from various nagios sources
type NagiosParserSvc interface {
	GetParsedNagios(ctx context.Context) (map[string][]parser.NagiosStatus, error)
	RefreshNagiosData(ctx context.Context) error
}

type nagiosParserSvc struct {
	statusDir string
	localDB   string
}

// NewNagiosParserSvc returns a boltdb backed nagios parser service
func NewNagiosParserSvc(statusDir, localDB string) (NagiosParserSvc, error) {
	svc := nagiosParserSvc{statusDir: statusDir, localDB: localDB}
	if _, err := os.Stat(statusDir); err != nil {
		return &svc, err
	}
	return &svc, nil
}

func openBoltDB(localDB string) (*bolt.DB, error) {
	db, err := bolt.Open(localDB, 0600, nil)
	if err != nil {
		return db, err
	}
	return db, nil
}

// GetParsedNagios returns a parsed list of nagios issues per host
func (svc *nagiosParserSvc) GetParsedNagios(ctx context.Context) (map[string][]parser.NagiosStatus, error) {
	result := make(map[string][]parser.NagiosStatus)
	localDB, err := openBoltDB(svc.localDB)
	if err != nil {
		return result, err
	}
	err = localDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("NagiosDB"))
		err = b.ForEach(func(k, v []byte) error {
			var statuses []parser.NagiosStatus
			err := json.Unmarshal(v, &statuses)
			if err != nil {
				return err
			}
			result[string(k)] = statuses
			return nil
		})
		return nil
	})
	localDB.Close()
	return result, err
}

//RefreshNagiosData returns a parsed map of hostname to issues from various nagios status files
func (svc *nagiosParserSvc) RefreshNagiosData(ctx context.Context) error {
	result := make(map[string][]parser.NagiosStatus)
	files, err := filepath.Glob(filepath.Join(svc.statusDir, "*.dat"))
	if err != nil {
		return err
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
		return fmt.Errorf("Failed to parse nagios data: %v ", <-errChan)
	}
	for resultChunk := range resultChan {
		for hostname, values := range resultChunk {
			result[hostname] = values
		}
	}
	// Marshall and Store results in localDB
	localDB, err := openBoltDB(svc.localDB)
	if err != nil {
		return err
	}
	err = localDB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("NagiosDB"))
		if err != nil {
			return err
		}
		for host, statuses := range result {
			statB, err := json.Marshal(statuses)
			if err != nil {
				return err
			}
			err = b.Put([]byte(host), statB)
		}
		return err
	})
	if err != nil {
		localDB.Close()
		return err
	}
	localDB.Close()
	return nil
}
