// (C) Copyright 2017 Hewlett Packard Enterprise Development LP
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sources

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func compareStatsMetrics(expectedMetrics []lustreStatsMetric, parsedMetric lustreStatsMetric) error {
	for _, metric := range expectedMetrics {
		if reflect.DeepEqual(metric, parsedMetric) {
			return nil
		}
	}
	return errors.New("Metric not found")
}

func TestRegexCaptureStrings(t *testing.T) {
	testString := `The lustre_exporter is a collector to be used with Prometheus which captures Lustre metrics.
Lustre is a parrallel filesystem for high-performance computers.
Currently, Lustre is on over 60% of the top supercomputers in the world.`
	// Matching is case-sensitive
	testPattern := "Lustre"
	expected := 3

	matchedStrings := regexCaptureStrings(testPattern, testString)
	if l := len(matchedStrings); l != expected {
		t.Fatalf("Retrieved an unexpected number of regex matches. Expected: %d, Got: %d", expected, l)
	}

	testPattern = "lustre"
	expected = 1

	matchedStrings = regexCaptureStrings(testPattern, testString)
	if l := len(matchedStrings); l != expected {
		t.Fatalf("Retrieved an unexpected number of regex matches. Expected: %d, Got: %d", expected, l)
	}

	// Matching is case-insensitive
	testPattern = "(?i)lustre"
	expected = 4

	matchedStrings = regexCaptureStrings(testPattern, testString)
	if l := len(matchedStrings); l != expected {
		t.Fatalf("Retrieved an unexpected number of regex matches. Expected: %d, Got: %d", expected, l)
	}

	// Match does not exist
	testPattern = "DNE"
	expected = 0

	matchedStrings = regexCaptureStrings(testPattern, testString)
	if l := len(matchedStrings); l != expected {
		t.Fatalf("Retrieved an unexpected number of regex matches. Expected: %d, Got: %d", expected, l)
	}
}

func TestRegexCaptureString(t *testing.T) {
	testString := "Hex Dump: 42 4F 49 4C 45 52 20 55 50"
	testPattern := "[0-9]*\\.[0-9]+|[0-9]+"
	expected := "42"

	matchedString := strings.TrimSpace(regexCaptureString(testPattern, testString))
	if matchedString != expected {
		t.Fatalf("Retrieved an unexpected string. Expected: %s, Got: %s", expected, matchedString)
	}

	testPattern = "DNE"
	expected = ""

	matchedString = strings.TrimSpace(regexCaptureString(testPattern, testString))
	if matchedString != expected {
		t.Fatalf("Retrieved an unexpected string. Expected: %s, Got: %s", expected, matchedString)
	}
}

func TestParseFileElements(t *testing.T) {
	testPath := "/proc/fs/lustre/obdfilter/OST0000/filesfree"
	directoryDepth := 0
	expectedName := "filesfree"
	expectedNodeName := "OST0000"

	name, nodeName, err := parseFileElements(testPath, directoryDepth)
	if err != nil {
		t.Fatal(err)
	}
	if name != expectedName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedName, name)
	}
	if nodeName != expectedNodeName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedNodeName, nodeName)
	}

	testPath = "/proc/fs/lustre/ldlm/namespaces/filter-lustrefs-OST0005_UUID/pool/grant_rate"
	directoryDepth = 1
	expectedName = "grant_rate"
	expectedNodeName = "lustrefs-OST0005"

	name, nodeName, err = parseFileElements(testPath, directoryDepth)
	if err != nil {
		t.Fatal(err)
	}
	if name != expectedName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedName, name)
	}
	if nodeName != expectedNodeName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedNodeName, nodeName)
	}

	testPath = "/proc/fs/lustre/health_check"
	directoryDepth = 0
	expectedName = "health_check"
	expectedNodeName = "lustre"

	name, nodeName, err = parseFileElements(testPath, directoryDepth)
	if err != nil {
		t.Fatal(err)
	}
	if name != expectedName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedName, name)
	}
	if nodeName != expectedNodeName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedNodeName, nodeName)
	}
}

func TestConvertToBytes(t *testing.T) {
	var resultString string
	testStringResults := map[string]string{
		"1":          "1",
		"512":        "512",
		"1K":         "1024",
		"10M":        "10485760",
		"100G":       "107374182400",
		"HotGarbage": "HotGarbage",
		"":           "",
		"KG":         "KG",
		"M":          "M",
		"100m":       "104857600",
	}

	for testString := range testStringResults {
		resultString = convertToBytes(testString)
		if resultString != testStringResults[testString] {
			t.Fatalf("Unexpected result of convertToBytes for %s: Expected %s, got %s", testString, testStringResults[testString], resultString)
		}
	}
}

func TestMultipleGetJobStats(t *testing.T) {
	testJobStatsBlock := `job_stats:
	- job_id:          24
	  snapshot_time:   1510782606
	  read_bytes:      { samples:         125, unit: bytes, min:    4096, max:    4096, sum:          512000 }
	  write_bytes:     { samples:       64575, unit: bytes, min:    4096, max: 4194304, sum:    215147593728 }
	  getattr:         { samples:           7, unit:  reqs }
	  setattr:         { samples:          43, unit:  reqs }
	  punch:           { samples:           1, unit:  reqs }
	  sync:            { samples:           8, unit:  reqs }
	  destroy:         { samples:          12, unit:  reqs }
	  create:          { samples:           5, unit:  reqs }
	  statfs:          { samples:           6, unit:  reqs }
	  get_info:        { samples:          23, unit:  reqs }
	  set_info:        { samples:          74, unit:  reqs }
	  quotactl:        { samples:           9, unit:  reqs }
	- job_id:          26
	  snapshot_time:   1510782606
	  read_bytes:      { samples:         125, unit: bytes, min:    4096, max:    4096, sum:          512000 }
	  write_bytes:     { samples:       56048, unit: bytes, min:    4096, max: 4194304, sum:    185838792704 }
	  getattr:         { samples:           7, unit:  reqs }
	  setattr:         { samples:          43, unit:  reqs }
	  punch:           { samples:           1, unit:  reqs }
	  sync:            { samples:           8, unit:  reqs }
	  destroy:         { samples:          12, unit:  reqs }
	  create:          { samples:           5, unit:  reqs }
	  statfs:          { samples:           6, unit:  reqs }
	  get_info:        { samples:          23, unit:  reqs }
	  set_info:        { samples:          74, unit:  reqs }
	  quotactl:        { samples:           9, unit:  reqs }
	- job_id:          28
	  snapshot_time:   1510782606
	  read_bytes:      { samples:         125, unit: bytes, min:    4096, max:    4096, sum:          512000 }
	  write_bytes:     { samples:       64208, unit: bytes, min:    4096, max: 4194304, sum:    213963751424 }
	  getattr:         { samples:           7, unit:  reqs }
	  setattr:         { samples:          43, unit:  reqs }
	  punch:           { samples:           1, unit:  reqs }
	  sync:            { samples:           8, unit:  reqs }
	  destroy:         { samples:          12, unit:  reqs }
	  create:          { samples:           5, unit:  reqs }
	  statfs:          { samples:           6, unit:  reqs }
	  get_info:        { samples:          23, unit:  reqs }
	  set_info:        { samples:          74, unit:  reqs }
	  quotactl:        { samples:           9, unit:  reqs }`
	expectedJobStats := 3

	capturedJobStats := regexCaptureStrings("(?ms:job_id:.*?$.*?(-|\\z))", testJobStatsBlock)
	if l := len(capturedJobStats); l != expectedJobStats {
		t.Fatalf("Retrieved an unexpected number of regex matches. Expected: %d, Got: %d", expectedJobStats, l)
	}

	exptectedLastJobStatEntry := "job_id:          28\n\t  snapshot_time:   1510782606\n\t  read_bytes:      { samples:         125, unit: bytes, min:    4096, max:    4096, sum:          512000 }\n\t  write_bytes:     { samples:       64208, unit: bytes, min:    4096, max: 4194304, sum:    213963751424 }\n\t  getattr:         { samples:           7, unit:  reqs }\n\t  setattr:         { samples:          43, unit:  reqs }\n\t  punch:           { samples:           1, unit:  reqs }\n\t  sync:            { samples:           8, unit:  reqs }\n\t  destroy:         { samples:          12, unit:  reqs }\n\t  create:          { samples:           5, unit:  reqs }\n\t  statfs:          { samples:           6, unit:  reqs }\n\t  get_info:        { samples:          23, unit:  reqs }\n\t  set_info:        { samples:          74, unit:  reqs }\n\t  quotactl:        { samples:           9, unit:  reqs }"

	if capturedJobStats[2] != exptectedLastJobStatEntry {
		t.Fatal("Comparision with last job stat entry failed.")
	}
}
