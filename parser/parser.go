package parser

import (
	"io/ioutil"
	"regexp"
	"strings"
)

// NagiosStatus is a status structure for nagios events
type NagiosStatus struct {
	StatusType string            `json:"status_type,omitempty"`
	Hostname   string            `json:"hostname,omitempty"`
	Values     map[string]string `json:"values,omitempty"`
}

func getRegExMap() (map[string]*regexp.Regexp, error) {
	reMap := make(map[string]*regexp.Regexp)
	var err error
	reMap["id"], err = regexp.Compile(`\s*(\w+)\s+{`)
	if err != nil {
		return reMap, err
	}
	reMap["attr"], err = regexp.Compile(`\s*(\w+)(?:=|\s+)(.*)`)
	if err != nil {
		return reMap, err
	}
	reMap["end"], err = regexp.Compile(`\s*}`)
	if err != nil {
		return reMap, err
	}
	return reMap, err
}

func getStateMapping() map[string]map[int]string {
	return map[string]map[int]string {
		"hosts":
		{
			0: "OK",
			1: "DOWN",
			2: "UNREACHABLE",
		},
		"services":
		{
			0: "OK",
			1: "WARNING",
			2: "CRITICAL",
			3: "UNKNOWN",
		},
	}

}

func newNagiosStatus() NagiosStatus {
	s := NagiosStatus{}
	s.Values = make(map[string]string)
	return s
}

// ParseStatusFromFile reads nagios entries from a file and returns a mapped listof issues per hostname
func ParseStatusFromFile(f string) (map[string][]NagiosStatus, error) {
	var result map[string][]NagiosStatus
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		return result, err
	}
	data := string(raw)
	return ParseStatus(data)
}

// ParseStatus parses status and returns a mapped list of issues per hostname
func ParseStatus(data string) (map[string][]NagiosStatus, error) {
	result := make(map[string][]NagiosStatus)
	lines := strings.Split(data, "\n")
	reMap, err := getRegExMap()
	if err != nil {
		return result, err
	}
	reID := reMap["id"]
	reAttr := reMap["attr"]
	reEnd := reMap["end"]
	cur := newNagiosStatus()
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) == 0 || l[0] == '#' {
			continue
		}
		if subMatch := reID.FindStringSubmatch(l); subMatch != nil {
			cur.StatusType = subMatch[1]
			continue
		}
		if subMatch := reAttr.FindStringSubmatch(l); subMatch != nil {
			key := subMatch[1]
			value := subMatch[2]
			if key == "host_name" {
				cur.Hostname = value
			} else {
				cur.Values[key] = value
			}
			continue
		}
		if matchID := reEnd.MatchString(l); matchID {
			result[cur.Hostname] = append(result[cur.Hostname], cur)
			cur = newNagiosStatus()
		}
	}
	return result, err
}
