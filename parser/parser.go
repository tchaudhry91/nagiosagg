package parser

import (
	"io/ioutil"
	"regexp"
	"strings"
)

// NagiosStatus is a status structure for nagios events
type NagiosStatus struct {
	statusType string
	hostname   string
	values     map[string]string
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

func getBlankNagiosStatus() NagiosStatus {
	s := NagiosStatus{}
	s.values = make(map[string]string)
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
	cur := getBlankNagiosStatus()
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if subMatch := reID.FindStringSubmatch(line); subMatch != nil {
			cur.statusType = subMatch[1]
			continue
		}
		if subMatch := reAttr.FindStringSubmatch(line); subMatch != nil {
			key := subMatch[1]
			value := subMatch[2]
			if key == "host_name" {
				cur.hostname = value
			} else {
				cur.values[key] = value
			}
			continue
		}
		if matchID := reEnd.MatchString(line); matchID {
			result[cur.hostname] = append(result[cur.hostname], cur)
			cur = getBlankNagiosStatus()
		}
	}
	return result, err
}
