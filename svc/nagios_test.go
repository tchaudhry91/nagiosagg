package svc

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var statusDir = flag.String("statusDir", "../samples/public", "Directory containing nagios .dat files")
var svc NagiosParserSvc

func init() {
	flag.Parse()
	svc, _ = NewNagiosParserSvc(*statusDir, filepath.Join(os.TempDir(), "tmp-test.boltdb"))
}

func TestNagiosData(t *testing.T) {
	ctx := context.TODO()
	t.Run("Populate", func(t *testing.T) {
		err := svc.RefreshNagiosData(ctx)
		if err != nil {
			t.Errorf("Population failed with: %v", err)
			t.FailNow()
		}
	})
	t.Run("Fetch", func(t *testing.T) {
		result, err := svc.GetParsedNagios(ctx)
		if err != nil {
			t.Errorf("Fetch of data failed with: %v", err)
			t.FailNow()
		}
		if len(result) < 3 {
			t.Errorf("Incorrect length of returned nagios list: %d", len(result))
			t.FailNow()
		}
	})
	os.Remove(filepath.Join(os.TempDir(), "tmp-test.boltdb"))
}
