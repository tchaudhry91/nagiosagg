package parser

import (
	"flag"
	"testing"
)

var nagiosFile = flag.String("nagiosfile", "../samples/public/status.dat", "sample nagios status.dat file to parse")

func TestRegularExpressions(t *testing.T) {
	reMap, err := getRegExMap()
	if err != nil {
		t.Errorf("Error compiling regular expressions")
		t.FailNow()
	}

	// Check reID
	reID := reMap["id"]
	line := "servicestatus {"
	if subMatch := reID.FindStringSubmatch(line); subMatch != nil {
		if subMatch[1] != "servicestatus" {
			t.Logf("wrong submatch for reID")
			t.Fail()
		}
	} else {
		t.Logf("nil Match on reID")
		t.Fail()
	}
	// Check reAttr
	reAttr := reMap["attr"]
	line = "host_name=yleoy-dev"
	if subMatch := reAttr.FindStringSubmatch(line); subMatch != nil {
		if subMatch[1] != "host_name" && subMatch[2] != "yleoy-dev" {
			t.Logf("wrong submatch for reAttr")
			t.Fail()
		}
	} else {
		t.Logf("nil Match on reAttr")
		t.Fail()
	}
	// Check reEnd
	reEnd := reMap["end"]
	line = "          }"
	if match := reEnd.MatchString(line); !match {
		t.Logf("Failed reEnd match")
		t.Fail()
	}
}

func TestParser(t *testing.T) {
	f := nagiosFile
	result, err := ParseStatusFromFile(*f)
	if err != nil {
		t.Errorf("Failed to parse nagios status:%v", err)
		t.FailNow()
	}
	if !(len(result) > 0) {
		t.Errorf("Invalid number of hosts in result: %d", len(result))
		t.FailNow()
	}
	var services, hosts, others int
	for _, r := range result {
		for _, status := range r {
			if status.StatusType == "hoststatus" {
				hosts++
				continue
			}
			if status.StatusType == "servicestatus" {
				services++
				continue
			}
			others++
		}
	}
	t.Logf("Following distribution was found:\nHostAlerts:%d\nServiceAlerts:%d\nOthers:%d", hosts, services, others)
	if (hosts+services) < 1 || others > 0 {
		t.Errorf("Following distribution was found:\nHostAlerts:%d\nServiceAlerts:%d\nOthers:%d", hosts, services, others)
		t.FailNow()
	}
}
