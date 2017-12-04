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

package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/HewlettPackard/lustre_exporter/sources"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
)

type labelPair struct {
	Name  string
	Value string
}

type promType struct {
	Name   string
	Help   string
	Type   int
	Labels []labelPair
	Value  float64
	Parsed bool // Parsed returns 'true' if the metric has already been parsed
}

const (
	// Constants taken from https://github.com/prometheus/client_model/blob/master/go/metrics.pb.go
	counter = 0
	gauge   = 1
	untyped = 3
)

var (
	errMetricNotFound      = errors.New("metric not found")
	errMetricAlreadyParsed = errors.New("metric already parsed")
)

func toggleCollectors(target string) {
	switch target {
	case "OST":
		sources.OstEnabled = "extended"
		sources.MdtEnabled = "disabled"
		sources.MgsEnabled = "disabled"
		sources.MdsEnabled = "disabled"
		sources.ClientEnabled = "disabled"
		sources.GenericEnabled = "disabled"
		sources.LnetEnabled = "disabled"
	case "MDT":
		sources.OstEnabled = "disabled"
		sources.MdtEnabled = "extended"
		sources.MgsEnabled = "disabled"
		sources.MdsEnabled = "disabled"
		sources.ClientEnabled = "disabled"
		sources.GenericEnabled = "disabled"
		sources.LnetEnabled = "disabled"
	case "MGS":
		sources.OstEnabled = "disabled"
		sources.MdtEnabled = "disabled"
		sources.MgsEnabled = "extended"
		sources.MdsEnabled = "disabled"
		sources.ClientEnabled = "disabled"
		sources.GenericEnabled = "disabled"
		sources.LnetEnabled = "disabled"
	case "MDS":
		sources.OstEnabled = "disabled"
		sources.MdtEnabled = "disabled"
		sources.MgsEnabled = "disabled"
		sources.MdsEnabled = "extended"
		sources.ClientEnabled = "disabled"
		sources.GenericEnabled = "disabled"
		sources.LnetEnabled = "disabled"
	case "Client":
		sources.OstEnabled = "disabled"
		sources.MdtEnabled = "disabled"
		sources.MgsEnabled = "disabled"
		sources.MdsEnabled = "disabled"
		sources.ClientEnabled = "extended"
		sources.GenericEnabled = "disabled"
		sources.LnetEnabled = "disabled"
	case "Generic":
		sources.OstEnabled = "disabled"
		sources.MdtEnabled = "disabled"
		sources.MgsEnabled = "disabled"
		sources.MdsEnabled = "disabled"
		sources.ClientEnabled = "disabled"
		sources.GenericEnabled = "extended"
		sources.LnetEnabled = "disabled"
	case "LNET":
		sources.OstEnabled = "disabled"
		sources.MdtEnabled = "disabled"
		sources.MgsEnabled = "disabled"
		sources.MdsEnabled = "disabled"
		sources.ClientEnabled = "disabled"
		sources.GenericEnabled = "disabled"
		sources.LnetEnabled = "extended"
	}
}

func stringAlphabetize(str1 string, str2 string) (int, error) {
	var letterCount int
	if len(str1) > len(str2) {
		letterCount = len(str2)
	} else {
		letterCount = len(str1)
	}

	for i := 0; i < letterCount-1; i++ {
		if str1[i] == str2[i] {
			continue
		} else if str1[i] > str2[i] {
			return 1, nil
		} else {
			return 2, nil
		}
	}

	return 0, fmt.Errorf("Duplicate label detected: %q", str1)
}

func sortByKey(labels []labelPair) ([]labelPair, error) {
	if len(labels) < 2 {
		return labels, nil
	}
	labelUpdate := make([]labelPair, len(labels))

	for i := range labels {
		desiredIndex := 0
		for j := range labels {
			if i == j {
				continue
			}

			result, err := stringAlphabetize(labels[i].Name, labels[j].Name)
			if err != nil {
				return nil, err
			}

			if result == 1 {
				//label for i's Name comes after label j's Name
				desiredIndex++
			}
		}
		labelUpdate[desiredIndex] = labels[i]
	}
	return labelUpdate, nil
}

func compareResults(parsedMetric promType, expectedMetrics []promType) ([]promType, error) {
	for i := range expectedMetrics {
		if reflect.DeepEqual(parsedMetric, expectedMetrics[i]) {
			if expectedMetrics[i].Parsed {
				return expectedMetrics, errMetricAlreadyParsed
			}
			expectedMetrics[i].Parsed = true
			return expectedMetrics, nil
		}
	}
	return expectedMetrics, errMetricNotFound
}

func blacklisted(blacklist []string, metricName string) bool {
	for _, name := range blacklist {
		if strings.HasPrefix(metricName, name) {
			return true
		}
	}
	return false
}

func TestCollector(t *testing.T) {
	targets := []string{"OST", "MDT", "MGS", "MDS", "Client", "Generic", "LNET"}
	// Override the default file location to the local proc directory
	sources.ProcLocation = "proc"

	expectedMetrics := []promType{
		// OST Metrics
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "connect"}, {"target", "lustrefs-OST0000"}, {"component", "ost"}}, 2, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "create"}, {"target", "lustrefs-OST0000"}, {"component", "ost"}}, 3, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "get_info"}, {"target", "lustrefs-OST0000"}, {"component", "ost"}}, 1, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "ping"}, {"target", "lustrefs-OST0000"}, {"component", "ost"}}, 14630, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "set_info_async"}, {"target", "lustrefs-OST0000"}, {"component", "ost"}}, 2, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "statfs"}, {"target", "lustrefs-OST0000"}, {"component", "ost"}}, 314187, false},
		{"lustre_grant_precreate_capacity_bytes", "Maximum space in bytes that clients can preallocate for objects", gauge, []labelPair{{"target", "lustrefs-OST0000"}, {"component", "ost"}}, 1.0174464e+07, false},
		{"lustre_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1, false},
		{"lustre_blocksize_bytes", "Filesystem block size in bytes", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_lfsck_speed_limit", "Maximum operations per second LFSCK (Lustre filesystem verification) can run", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_shrink_requests_total", "Number of shrinks that have been requested", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_exports_pending_total", "Total number of exports that have been marked pending", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 2.513043456e+10, false},
		{"lustre_recalc_freed_total", "Number of locks that have been freed", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_lock_contended_total", "Number of contended locks", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 32, false},
		{"lustre_recovery_time_hard_seconds", "Maximum timeout 'recover_time_soft' can increment to for a single server", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 900, false},
		{"lustre_lock_timeout_total", "Number of lock timeouts", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_locks_grant_total", "Total number of granted locks", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_locks_granted", "Number of granted less cancelled locks", untyped, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 20, false},
		{"lustre_brw_size_megabytes", "Block read/write size in megabytes", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1, false},
		{"lustre_exports_dirty_total", "Total number of exports that have been marked dirty", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 3.555328e+07, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 2.08901812224e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 2.06254907392e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 2.06309617664e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 2.09064738816e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 2.095316992e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.11325073408e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 9.4739140608e+10, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 9.4455115776e+10, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.0474225664e+10, false},
		{"lustre_lock_count_total", "Number of locks", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 20, false},
		{"lustre_server_lock_volume", "Current value for server lock volume (SLV)", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 2.315574e+11, false},
		{"lustre_shrink_freed_total", "Number of shrinks that have been freed", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_available_kilobytes", "Number of kilobytes readily available in the pool", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.54299648e+10, false},
		{"lustre_exports_total", "Total number of times the pool has been exported", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 3, false},
		{"lustre_lock_cancel_rate", "Lock cancel rate", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_lock_grant_plan", "Number of planned lock grants per second", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 64341, false},
		{"lustre_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_sync_journal_enabled", "Binary indicator as to whether or not the journal is set for asynchronous commits", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_lock_cancel_total", "Total number of cancelled locks", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_recovery_time_soft_seconds", "Duration in seconds for a client to attempt to reconnect after a crash (automatically incremented if servers are still in an error state)", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 300, false},
		{"lustre_soft_sync_limit", "Number of RPCs necessary before triggering a sync", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 16, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 11, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 9, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 10, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_exports_granted_total", "Total number of exports that have been marked granted", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 2.5218236416e+10, false},
		{"lustre_job_cleanup_interval_seconds", "Interval in seconds between cleanup of tuning statistics", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 600, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_lock_contention_seconds_total", "Time in seconds during which locks were contended", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 2, false},
		{"lustre_precreate_batch", "Maximum number of objects that can be included in a single transaction", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 128, false},
		{"lustre_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 5.1473433403392e+13, false},
		{"lustre_free_kilobytes", "Number of kilobytes allocated to the pool", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.5454542848e+10, false},
		{"lustre_inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.5092572e+07, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 199575, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 197501, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 198901, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 199384, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 199832, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 106178, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 92053, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 91140, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 9991, false},
		{"lustre_inodes_free", "The number of inodes (objects) available", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.5092327e+07, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 45056, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 327680, false},
		{"lustre_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_degraded", "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_grant_compat_disabled", "Binary indicator as to whether clients with OBD_CONNECT_GRANT_PARAM setting will be granted space", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_recalc_timing_seconds_total", "Number of seconds spent locked", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_capacity_kilobytes", "Capacity of the pool in kilobytes", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.5722801152e+10, false},
		{"lustre_lock_grant_rate", "Lock grant rate", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 0, false},
		{"lustre_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 1.048576e+06, false},
		{"lustre_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 4096, false},
		{"lustre_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0000"}}, 4.9185469e+07, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1"}}, 5, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "128"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "16"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "2"}}, 2, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "256"}}, 1, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "32"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "4"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "64"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "8"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1"}}, 6619, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "128"}}, 52859, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "16"}}, 6843, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "2"}}, 1078, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "256"}}, 6.5974555e+07, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "32"}}, 13163, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "4"}}, 1884, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "64"}}, 26868, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "8"}}, 3648, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "0"}}, 8, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "10"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "11"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "12"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "13"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "14"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "15"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "16"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "17"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "18"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "19"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "2"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "20"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "21"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "22"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "23"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "24"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "25"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "26"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "27"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "28"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "29"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "3"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "30"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "31"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "4"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "5"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "6"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "7"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "8"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "9"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "0"}}, 6619, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1"}}, 1078, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "10"}}, 872, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "11"}}, 891, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "12"}}, 850, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "13"}}, 881, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "14"}}, 813, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "15"}}, 829, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "16"}}, 840, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "17"}}, 887, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "18"}}, 821, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "19"}}, 815, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "2"}}, 923, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "20"}}, 805, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "21"}}, 805, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "22"}}, 847, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "23"}}, 843, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "24"}}, 779, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "25"}}, 863, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "26"}}, 777, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "27"}}, 828, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "28"}}, 859, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "29"}}, 801, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "3"}}, 961, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "30"}}, 771, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "31"}}, 6.6055104e+07, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "4"}}, 930, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "5"}}, 932, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "6"}}, 885, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "7"}}, 901, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "8"}}, 876, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "9"}}, 831, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1024"}}, 1, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1048576"}}, 1, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "131072"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "16384"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "2048"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "262144"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "32768"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "4096"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "524288"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "65536"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "8192"}}, 3, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1024"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1048576"}}, 6.5974555e+07, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "131072"}}, 13163, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "16384"}}, 1884, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "2048"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "262144"}}, 26868, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "32768"}}, 3648, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "4096"}}, 6619, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "524288"}}, 52859, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "65536"}}, 6843, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "8192"}}, 1078, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1"}}, 8, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "10"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "11"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "12"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "13"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "14"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "2"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "3"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "4"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "5"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "6"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "7"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "8"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "9"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "1"}}, 3.9683886e+07, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "10"}}, 10445, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "11"}}, 1172, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "12"}}, 254, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "13"}}, 47, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "14"}}, 10, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "2"}}, 4.204676e+06, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "3"}}, 768865, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "4"}}, 552287, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "5"}}, 935658, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "6"}}, 3.197433e+06, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "7"}}, 1.0469836e+07, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "8"}}, 5.745458e+06, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0000"}, {"size", "9"}}, 517490, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "connect"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 3, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "ping"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 14630, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "set_info_async"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 314187, false},
		{"lustre_grant_precreate_capacity_bytes", "Maximum space in bytes that clients can preallocate for objects", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.0174464e+07, false},
		{"lustre_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1, false},
		{"lustre_blocksize_bytes", "Filesystem block size in bytes", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_lfsck_speed_limit", "Maximum operations per second LFSCK (Lustre filesystem verification) can run", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_shrink_requests_total", "Number of shrinks that have been requested", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_exports_pending_total", "Total number of exports that have been marked pending", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2.513043456e+10, false},
		{"lustre_recalc_freed_total", "Number of locks that have been freed", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_lock_contended_total", "Number of contended locks", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 32, false},
		{"lustre_recovery_time_hard_seconds", "Maximum timeout 'recover_time_soft' can increment to for a single server", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 900, false},
		{"lustre_lock_timeout_total", "Number of lock timeouts", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_locks_grant_total", "Total number of granted locks", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_locks_granted", "Number of granted less cancelled locks", untyped, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 20, false},
		{"lustre_brw_size_megabytes", "Block read/write size in megabytes", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1, false},
		{"lustre_exports_dirty_total", "Total number of exports that have been marked dirty", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 3.555328e+07, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2.08901812224e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2.06254907392e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2.06309617664e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2.09064738816e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2.095316992e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.11325073408e+11, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 9.4739140608e+10, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 9.4455115776e+10, false},
		{"lustre_job_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.0474225664e+10, false},
		{"lustre_lock_count_total", "Number of locks", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 20, false},
		{"lustre_server_lock_volume", "Current value for server lock volume (SLV)", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2.315574e+11, false},
		{"lustre_shrink_freed_total", "Number of shrinks that have been freed", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_available_kilobytes", "Number of kilobytes readily available in the pool", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.54299648e+10, false},
		{"lustre_exports_total", "Total number of times the pool has been exported", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 3, false},
		{"lustre_lock_cancel_rate", "Lock cancel rate", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_lock_grant_plan", "Number of planned lock grants per second", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 64341, false},
		{"lustre_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_sync_journal_enabled", "Binary indicator as to whether or not the journal is set for asynchronous commits", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_lock_cancel_total", "Total number of cancelled locks", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_recovery_time_soft_seconds", "Duration in seconds for a client to attempt to reconnect after a crash (automatically incremented if servers are still in an error state)", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 300, false},
		{"lustre_soft_sync_limit", "Number of RPCs necessary before triggering a sync", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 16, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "334"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "335"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "336"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "337"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "338"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 11, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "339"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 9, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "340"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 10, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "341"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "create"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "destroy"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "get_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "getattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "punch"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "quotactl"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "set_info"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "setattr"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "statfs"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "342"}, {"operation", "sync"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_exports_granted_total", "Total number of exports that have been marked granted", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2.5218236416e+10, false},
		{"lustre_job_cleanup_interval_seconds", "Interval in seconds between cleanup of tuning statistics", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 600, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_lock_contention_seconds_total", "Time in seconds during which locks were contended", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 2, false},
		{"lustre_precreate_batch", "Maximum number of objects that can be included in a single transaction", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 128, false},
		{"lustre_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 5.1473433403392e+13, false},
		{"lustre_free_kilobytes", "Number of kilobytes allocated to the pool", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.5454542848e+10, false},
		{"lustre_inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.5092572e+07, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 199575, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 197501, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 198901, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 199384, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 199832, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 106178, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 92053, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 91140, false},
		{"lustre_job_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 9991, false},
		{"lustre_inodes_free", "The number of inodes (objects) available", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.5092327e+07, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 45056, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 4096, false},
		{"lustre_job_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 327680, false},
		{"lustre_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_degraded", "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_grant_compat_disabled", "Binary indicator as to whether clients with OBD_CONNECT_GRANT_PARAM setting will be granted space", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "334"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "335"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "336"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "337"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "338"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "339"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "340"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "341"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_job_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"jobid", "342"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_recalc_timing_seconds_total", "Number of seconds spent locked", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_capacity_kilobytes", "Capacity of the pool in kilobytes", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.5722801152e+10, false},
		{"lustre_lock_grant_rate", "Lock grant rate", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 0, false},
		{"lustre_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 1.048576e+06, false},
		{"lustre_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 4096, false},
		{"lustre_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"component", "ost"}, {"target", "lustrefs-OST0001"}}, 4.9185469e+07, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1"}}, 5, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "128"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "16"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "2"}}, 2, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "256"}}, 1, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "32"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "4"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "64"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "8"}}, 0, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1"}}, 6619, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "128"}}, 52859, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "16"}}, 6843, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "2"}}, 1078, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "256"}}, 6.5974555e+07, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "32"}}, 13163, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "4"}}, 1884, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "64"}}, 26868, false},
		{"lustre_pages_per_bulk_rw_total", "Total number of pages per block RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "8"}}, 3648, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "0"}}, 8, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "10"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "11"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "12"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "13"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "14"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "15"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "16"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "17"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "18"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "19"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "2"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "20"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "21"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "22"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "23"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "24"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "25"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "26"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "27"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "28"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "29"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "3"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "30"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "31"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "4"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "5"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "6"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "7"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "8"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "9"}}, 0, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "0"}}, 6619, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1"}}, 1078, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "10"}}, 872, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "11"}}, 891, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "12"}}, 850, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "13"}}, 881, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "14"}}, 813, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "15"}}, 829, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "16"}}, 840, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "17"}}, 887, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "18"}}, 821, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "19"}}, 815, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "2"}}, 923, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "20"}}, 805, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "21"}}, 805, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "22"}}, 847, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "23"}}, 843, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "24"}}, 779, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "25"}}, 863, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "26"}}, 777, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "27"}}, 828, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "28"}}, 859, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "29"}}, 801, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "3"}}, 961, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "30"}}, 771, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "31"}}, 6.6055104e+07, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "4"}}, 930, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "5"}}, 932, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "6"}}, 885, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "7"}}, 901, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "8"}}, 876, false},
		{"lustre_discontiguous_pages_total", "Total number of logical discontinuities per RPC.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "9"}}, 831, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1024"}}, 1, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1048576"}}, 1, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "131072"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "16384"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "2048"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "262144"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "32768"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "4096"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "524288"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "65536"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "8192"}}, 3, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1024"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1048576"}}, 6.5974555e+07, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "131072"}}, 13163, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "16384"}}, 1884, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "2048"}}, 0, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "262144"}}, 26868, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "32768"}}, 3648, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "4096"}}, 6619, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "524288"}}, 52859, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "65536"}}, 6843, false},
		{"lustre_disk_io_total", "Total number of operations the filesystem has performed for the given size.", counter, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "8192"}}, 1078, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1"}}, 8, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "10"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "11"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "12"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "13"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "14"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "2"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "3"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "4"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "5"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "6"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "7"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "8"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "read"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "9"}}, 0, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "1"}}, 3.9683886e+07, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "10"}}, 10445, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "11"}}, 1172, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "12"}}, 254, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "13"}}, 47, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "14"}}, 10, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "2"}}, 4.204676e+06, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "3"}}, 768865, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "4"}}, 552287, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "5"}}, 935658, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "6"}}, 3.197433e+06, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "7"}}, 1.0469836e+07, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "8"}}, 5.745458e+06, false},
		{"lustre_disk_io", "Current number of I/O operations that are processing during the snapshot.", gauge, []labelPair{{"operation", "write"}, {"component", "ost"}, {"target", "lustrefs-OST0001"}, {"size", "9"}}, 517490, false},

		// MDT Metrics
		{"lustre_exports_total", "Total number of times the pool has been exported", counter, []labelPair{{"component", "mdt"}, {"target", "lustrefs-MDT0000"}}, 5, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 9064, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 9154, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 2, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 9066, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 9958, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 57434, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 20, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 33, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "crossdir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 38, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "link"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mkdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mknod"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rmdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "samedir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "sync"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "343"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 31, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "crossdir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 38, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "link"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mkdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mknod"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 33, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rmdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "samedir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "sync"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "344"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "crossdir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 39, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "link"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mkdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mknod"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 31, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rmdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "samedir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "sync"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "345"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "crossdir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 29, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "link"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mkdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mknod"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rmdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "samedir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "sync"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "346"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "crossdir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 31, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "link"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mkdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mknod"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rmdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "samedir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 40, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "sync"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "347"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 10, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "crossdir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 37, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "link"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mkdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mknod"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 38, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rmdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "samedir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 38, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "sync"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "348"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 35, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "crossdir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 30, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "link"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mkdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mknod"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 10, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rmdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "samedir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 38, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "sync"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "349"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 29, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "crossdir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 32, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "link"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mkdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mknod"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 27, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rmdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "samedir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 38, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "sync"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "350"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "close"}}, 23, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "crossdir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getattr"}}, 12, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "getxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "link"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mkdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "mknod"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "open"}}, 24, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "rmdir"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "samedir_rename"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setattr"}}, 24, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "setxattr"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "statfs"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "sync"}}, 0, false},
		{"lustre_job_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"jobid", "351"}, {"component", "mdt"}, {"target", "lustrefs-MDT0000"}, {"operation", "unlink"}}, 0, false},

		// MGS Metrics
		{"lustre_available_kilobytes", "Number of kilobytes readily available in the pool", gauge, []labelPair{{"target", "osd"}, {"component", "mgs"}}, 1.120742144e+09, false},
		{"lustre_blocksize_bytes", "Filesystem block size in bytes", gauge, []labelPair{{"component", "mgs"}, {"target", "osd"}}, 131072, false},
		{"lustre_capacity_kilobytes", "Capacity of the pool in kilobytes", gauge, []labelPair{{"component", "mgs"}, {"target", "osd"}}, 1.120748032e+09, false},
		{"lustre_quota_iused_estimate", "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated", gauge, []labelPair{{"component", "mgs"}, {"target", "osd"}}, 0, false},
		{"lustre_inodes_free", "The number of inodes (objects) available", gauge, []labelPair{{"component", "mgs"}, {"target", "osd"}}, 3.5980923e+07, false},
		{"lustre_inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gauge, []labelPair{{"component", "mgs"}, {"target", "osd"}}, 3.5981071e+07, false},
		{"lustre_free_kilobytes", "Number of kilobytes allocated to the pool", gauge, []labelPair{{"component", "mgs"}, {"target", "osd"}}, 1.120744192e+09, false},

		// MDS Metrics
		{"lustre_blocksize_bytes", "Filesystem block size in bytes", gauge, []labelPair{{"component", "mds"}, {"target", "osd"}}, 131072, false},
		{"lustre_capacity_kilobytes", "Capacity of the pool in kilobytes", gauge, []labelPair{{"component", "mds"}, {"target", "osd"}}, 1.120748032e+09, false},
		{"lustre_quota_iused_estimate", "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated", gauge, []labelPair{{"component", "mds"}, {"target", "osd"}}, 0, false},
		{"lustre_inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gauge, []labelPair{{"component", "mds"}, {"target", "osd"}}, 3.5981071e+07, false},
		{"lustre_available_kilobytes", "Number of kilobytes readily available in the pool", gauge, []labelPair{{"component", "mds"}, {"target", "osd"}}, 1.120742144e+09, false},
		{"lustre_inodes_free", "The number of inodes (objects) available", gauge, []labelPair{{"component", "mds"}, {"target", "osd"}}, 3.5980923e+07, false},
		{"lustre_free_kilobytes", "Number of kilobytes allocated to the pool", gauge, []labelPair{{"component", "mds"}, {"target", "osd"}}, 1.120744192e+09, false},

		// Client Metrics
		{"lustre_blocksize_bytes", "Filesystem block size in bytes", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1.048576e+06, false},
		{"lustre_checksum_pages_enabled", "Returns '1' if data checksumming is enabled for the client", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1, false},
		{"lustre_default_ea_size_bytes", "Default Extended Attribute (EA) size in bytes", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 128, false},
		{"lustre_inodes_free", "The number of inodes (objects) available", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 3.0171029e+07, false},
		{"lustre_inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 3.0171285e+07, false},
		{"lustre_available_kilobytes", "Number of kilobytes readily available in the pool", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 3.0869986304e+10, false},
		{"lustre_free_kilobytes", "Number of kilobytes allocated to the pool", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 3.0895133696e+10, false},
		{"lustre_capacity_kilobytes", "Capacity of the pool in kilobytes", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 3.1445602304e+10, false},
		{"lustre_lazystatfs_enabled", "Returns '1' if lazystatfs (a non-blocking alternative to statfs) is enabled for the client", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1, false},
		{"lustre_maximum_ea_size_bytes", "Maximum Extended Attribute (EA) size in bytes", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 128, false},
		{"lustre_maximum_read_ahead_megabytes", "Maximum number of megabytes to read ahead", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 64, false},
		{"lustre_maximum_read_ahead_per_file_megabytes", "Maximum number of megabytes per file to read ahead", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 64, false},
		{"lustre_maximum_read_ahead_whole_megabytes", "Maximum file size in megabytes for a file to be read in its entirety", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 2, false},
		{"lustre_statahead_agl_enabled", "Returns '1' if the Asynchronous Glimpse Lock (AGL) for statahead is enabled", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 0, false},
		{"lustre_statahead_maximum", "Maximum window size for statahead", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 0, false},
		{"lustre_read_samples_total", "Total number of reads that have been recorded.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 46318, false},
		{"lustre_read_minimum_size_bytes", "The minimum read size in bytes.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1.048576e+06, false},
		{"lustre_read_maximum_size_bytes", "The maximum read size in bytes.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1.048576e+06, false},
		{"lustre_read_bytes_total", "The total number of bytes that have been read.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 4.8567943168e+10, false},
		{"lustre_write_samples_total", "Total number of writes that have been recorded.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1.02484615e+08, false},
		{"lustre_write_minimum_size_bytes", "The minimum write size in bytes.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1.048576e+06, false},
		{"lustre_write_maximum_size_bytes", "The maximum write size in bytes.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1.048576e+06, false},
		{"lustre_write_bytes_total", "The total number of bytes that have been written.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1.0746290765824e+14, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "alloc_inode"}}, 80, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "close"}}, 10267, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "getattr"}}, 224, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "getxattr"}}, 1.02495e+08, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "inode_permission"}}, 21245, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "open"}}, 10271, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "readdir"}}, 30, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "removexattr"}}, 10171, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "seek"}}, 20, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "statfs"}}, 57491, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "truncate"}}, 10171, false},
		{"lustre_stats_total", "Number of operations the filesystem has performed.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}, {"operation", "unlink"}}, 20, false},
		{"lustre_xattr_cache_enabled", "Returns '1' if extended attribute cache is enabled", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-ffff88082d691000"}}, 1, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-MDT0000-mdc-ffff88082d691000"}, {"operation", "read"}, {"size", "0"}, {"type", "mdc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-MDT0000-mdc-ffff88082d691000"}, {"operation", "read"}, {"size", "1"}, {"type", "mdc"}}, 29659, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-MDT0000-mdc-ffff88082d691000"}, {"operation", "read"}, {"size", "2"}, {"type", "mdc"}}, 94, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-MDT0000-mdc-ffff88082d691000"}, {"operation", "read"}, {"size", "3"}, {"type", "mdc"}}, 2, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "0"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1"}, {"type", "osc"}}, 1, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "10"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "11"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "12"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "13"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "14"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "15"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "16"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "3"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "4"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "5"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "6"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "7"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "8"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "9"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "0"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1"}, {"type", "osc"}}, 797536, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "10"}, {"type", "osc"}}, 5.43314e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "11"}, {"type", "osc"}}, 1.130772e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "12"}, {"type", "osc"}}, 116573, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "13"}, {"type", "osc"}}, 13419, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "14"}, {"type", "osc"}}, 2122, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "15"}, {"type", "osc"}}, 253, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "16"}, {"type", "osc"}}, 16, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2"}, {"type", "osc"}}, 1.1977376e+07, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "3"}, {"type", "osc"}}, 8.395435e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "4"}, {"type", "osc"}}, 6.551231e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "5"}, {"type", "osc"}}, 3.483172e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "6"}, {"type", "osc"}}, 1.971052e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "7"}, {"type", "osc"}}, 1.245475e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "8"}, {"type", "osc"}}, 2.919367e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "9"}, {"type", "osc"}}, 8.773623e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "0"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1"}, {"type", "osc"}}, 1, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "10"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "11"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "12"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "13"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "14"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "15"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "16"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "17"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "3"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "4"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "5"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "6"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "7"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "8"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "9"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "0"}, {"type", "osc"}}, 0, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1"}, {"type", "osc"}}, 1.272532e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "10"}, {"type", "osc"}}, 4.475746e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "11"}, {"type", "osc"}}, 938831, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "12"}, {"type", "osc"}}, 92950, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "13"}, {"type", "osc"}}, 10449, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "14"}, {"type", "osc"}}, 1560, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "15"}, {"type", "osc"}}, 185, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "16"}, {"type", "osc"}}, 13, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "17"}, {"type", "osc"}}, 1, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2"}, {"type", "osc"}}, 1.3096231e+07, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "3"}, {"type", "osc"}}, 8.701266e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "4"}, {"type", "osc"}}, 6.35365e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "5"}, {"type", "osc"}}, 3.202186e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "6"}, {"type", "osc"}}, 1.694121e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "7"}, {"type", "osc"}}, 1.013376e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "8"}, {"type", "osc"}}, 2.447018e+06, false},
		{"lustre_rpcs_in_flight", "Current number of RPCs that are processing during the snapshot.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "9"}, {"type", "osc"}}, 7.139528e+06, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "0"}}, 1, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1024"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1048576"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "128"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "131072"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "16"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "16384"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2048"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2097152"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "256"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "262144"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "32"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "32768"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "4"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "4096"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "512"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "524288"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "64"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "65536"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "8"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "8192"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "0"}}, 5261, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1024"}}, 21122, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1048576"}}, 2.162882e+07, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "128"}}, 13, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "131072"}}, 2.707283e+06, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "16"}}, 1, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "16384"}}, 338082, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2048"}}, 42248, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2097152"}}, 9.544005e+06, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "256"}}, 5278, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "262144"}}, 5.408781e+06, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "32"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "32768"}}, 676679, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "4"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "4096"}}, 84502, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "512"}}, 10562, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "524288"}}, 1.081536e+07, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "64"}}, 6, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "65536"}}, 1.353553e+06, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "8"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "8192"}}, 169006, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "0"}}, 1, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1024"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1048576"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "128"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "131072"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "16"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "16384"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2048"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2097152"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "256"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "262144"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "32"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "32768"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "4"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "4096"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "512"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "524288"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "64"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "65536"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "8"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "8192"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "0"}}, 5021, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1024"}}, 20175, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1048576"}}, 2.0656694e+07, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "128"}}, 7, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "131072"}}, 2.585278e+06, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "16"}}, 2, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "16384"}}, 322949, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2048"}}, 40335, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2097152"}}, 9.115824e+06, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "256"}}, 5042, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "262144"}}, 5.165585e+06, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "32"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "32768"}}, 645946, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "4"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "4096"}}, 80718, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "512"}}, 10092, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "524288"}}, 1.0331924e+07, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "64"}}, 7, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "65536"}}, 1.292621e+06, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "8"}}, 0, false},
		{"lustre_rpcs_offset", "Current RPC offset by size.", gauge, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "8192"}}, 161423, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "128"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "16"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "256"}}, 1, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "32"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "4"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "64"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "8"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1"}}, 6156, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "128"}}, 50882, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "16"}}, 6526, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2"}}, 1032, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "256"}}, 5.270231e+07, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "32"}}, 12622, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "4"}}, 1798, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "64"}}, 25744, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0000-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "8"}}, 3492, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "1"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "128"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "16"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "2"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "256"}}, 1, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "32"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "4"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "64"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "read"}, {"size", "8"}}, 0, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "1"}}, 5298, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "128"}}, 47202, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "16"}}, 5897, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "2"}}, 931, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "256"}}, 5.0340888e+07, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "32"}}, 11296, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "4"}}, 1667, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "64"}}, 23471, false},
		{"lustre_pages_per_rpc_total", "Total number of pages per RPC.", counter, []labelPair{{"component", "client"}, {"target", "lustrefs-OST0001-osc-ffff88082d691000"}, {"operation", "write"}, {"size", "8"}}, 2993, false},

		// Generic Metrics
		{"lustre_health_check", "Current health status for the indicated instance: 1 refers to 'healthy', 0 refers to 'unhealthy'", gauge, []labelPair{{"component", "generic"}, {"target", "lustre"}}, 1, false},
		{"lustre_cache_miss_total", "Total number of cache misses.", counter, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_cache_access_total", "Total number of times cache has been accessed.", counter, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_free_pages", "Current number of pages available.", gauge, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_maximum_pools", "Number of pools.", gauge, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 8041, false},
		{"lustre_pages_in_pools", "Number of pages in all pools.", gauge, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_pages_per_pool", "Number of pages per pool.", gauge, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 512, false},
		{"lustre_maximum_pages_reached_total", "Total number of pages reached.", counter, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_maximum_pages", "Maximum number of pages that can be held.", gauge, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 4.116587e+06, false},
		{"lustre_physical_pages", "Capacity of physical memory.", gauge, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 3.2932703e+07, false},
		{"lustre_grows_total", "Total number of grows.", counter, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_maximum_waitqueue_depth", "Maximum waitqueue length.", gauge, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_grows_failure_total", "Total number of failures while attempting to add pages.", counter, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_shrinks_total", "Total number of shrinks.", counter, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_free_page_low", "Lowest number of free pages reached.", gauge, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},
		{"lustre_out_of_memory_request_total", "Total number of out of memory requests.", 0, []labelPair{{"component", "generic"}, {"target", "sptlrpc"}}, 0, false},

		// LNET Metrics
		{"lustre_console_max_delay_centiseconds", "Minimum time in centiseconds before the console logs a message", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 60000, false},
		{"lustre_drop_bytes_total", "Total number of bytes that have been dropped", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 0, false},
		{"lustre_maximum", "Maximum number of outstanding messages", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 28, false},
		{"lustre_route_count_total", "Total number of messages that have been routed", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 0, false},
		{"lustre_console_backoff_enabled", "Returns non-zero number if console_backoff is enabled", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 2, false},
		{"lustre_receive_count_total", "Total number of messages that have been received", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 1.01719291e+08, false},
		{"lustre_console_min_delay_centiseconds", "Maximum time in centiseconds before the console logs a message", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 50, false},
		{"lustre_receive_bytes_total", "Total number of bytes received", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 5.302956535372e+13, false},
		{"lustre_send_count_total", "Total number of messages that have been sent", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 1.01719323e+08, false},
		{"lustre_allocated", "Number of messages currently allocated", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 0, false},
		{"lustre_catastrophe_enabled", "Returns 1 if currently in catastrophe mode", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 0, false},
		{"lustre_errors_total", "Total number of errors", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 0, false},
		{"lustre_lnet_memory_used_bytes", "Number of bytes allocated by LNET", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 3.6109496e+07, false},
		{"lustre_send_bytes_total", "Total number of bytes sent", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 2.1201322992e+10, false},
		{"lustre_route_bytes_total", "Total number of bytes for routed messages", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 0, false},
		{"lustre_fail_error_total", "Number of errors that have been thrown", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 0, false},
		{"lustre_watchdog_ratelimit_enabled", "Returns 1 if the watchdog rate limiter is enabled", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 300, false},
		{"lustre_console_ratelimit_enabled", "Returns 1 if the console message rate limiting is enabled", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 1, false},
		{"lustre_debug_megabytes", "Maximum buffer size in megabytes for the LNET debug messages", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 101, false},
		{"lustre_drop_count_total", "Total number of messages that have been dropped", counter, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 0, false},
		{"lustre_fail_maximum", "Maximum number of times to fail", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 0, false},
		{"lustre_panic_on_lbug_enabled", "Returns 1 if panic_on_lbug is enabled", gauge, []labelPair{{"component", "lnet"}, {"target", "lnet"}}, 1, false},
	}

	// These following metrics should be filtered out as they are specific to the deployment and will always change
	blacklistedMetrics := []string{"go_", "http_", "process_", "lustre_exporter_"}

	for i, metric := range expectedMetrics {
		newLabels, err := sortByKey(metric.Labels)
		if err != nil {
			t.Fatal(err)
		}
		expectedMetrics[i].Labels = newLabels
	}

	numParsed := 0
	for _, target := range targets {
		toggleCollectors(target)
		var missingMetrics []promType // Array of metrics that are missing for the given target
		enabledSources := []string{"procfs", "procsys"}

		sourceList, err := loadSources(enabledSources)
		if err != nil {
			t.Fatal("Unable to load sources")
		}
		if err = prometheus.Register(LustreSource{sourceList: sourceList}); err != nil {
			t.Fatalf("Failed to register for target: %s", target)
		}

		promServer := httptest.NewServer(promhttp.Handler())
		defer promServer.Close()

		resp, err := http.Get(promServer.URL)
		if err != nil {
			t.Fatalf("Failed to GET data from prometheus: %v", err)
		}

		var parser expfmt.TextParser

		metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		for _, metricFamily := range metricFamilies {
			if blacklisted(blacklistedMetrics, *metricFamily.Name) {
				continue
			}
			for _, metric := range metricFamily.Metric {
				var value float64
				if *metricFamily.Type == counter {
					value = *metric.Counter.Value
				} else if *metricFamily.Type == gauge {
					value = *metric.Gauge.Value
				} else if *metricFamily.Type == untyped {
					value = *metric.Untyped.Value
				}
				var labels []labelPair
				for _, label := range metric.Label {
					l := labelPair{
						Name:  *label.Name,
						Value: *label.Value,
					}
					labels = append(labels, l)
				}
				p := promType{
					Name:   *metricFamily.Name,
					Help:   *metricFamily.Help,
					Type:   int(*metricFamily.Type),
					Labels: labels,
					Value:  value,
				}

				// Check if exists here
				expectedMetrics, err = compareResults(p, expectedMetrics)
				if err == errMetricNotFound {
					missingMetrics = append(missingMetrics, p)
				} else if err == errMetricAlreadyParsed {
					t.Fatalf("Retrieved an unexpected duplicate of %s metric: %+v", target, p)
				}
				numParsed++
			}
		}
		if len(missingMetrics) != 0 {
			t.Fatalf("The following %s metrics were not found: %+v", target, missingMetrics)
		}

		prometheus.Unregister(LustreSource{sourceList: sourceList})
	}

	if l := len(expectedMetrics); l != numParsed {
		t.Fatalf("Retrieved an unexpected number of metrics. Expected: %d, Got: %d", l, numParsed)
	}

	// Return the proc location to the default value
	sources.ProcLocation = "/proc"
}