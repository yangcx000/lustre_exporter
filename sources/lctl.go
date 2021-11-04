// (C) Copyright 2021 Gabriele Iannetti <g.iannetti@gsi.de>
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
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	lctlCmdCall                       = "" // If user not root use sudo as command, otherwise lctl
	lctlCmdArgs                       = []string{}
	changelogTargetRegexPattern       = regexp.MustCompile(`mdd.([\w\d-]+-MDT[\d]+).changelog_users=`)
	changelogCurrentIndexRegexPattern = regexp.MustCompile(`current index: (\d+)`)
	changelogUserRegexPattern         = regexp.MustCompile(`(?ms:(cl\d+)\s+(\d+) \((\d+)\))`)
)

const (
	changelogUsersCmdArg = "mdd.*-*.changelog_users"
)

type lustreLctlMetricCreator struct {
	argument      string
	metricHandler func(string) ([]prometheus.Metric, error)
}

func init() {
	Factories["lctl"] = newLustreLctlSource
}

func regexCaptureChangelogTarget(textToMatch string) (string, error) {
	matchedTarget := changelogTargetRegexPattern.FindStringSubmatch(textToMatch)
	if matchedTarget != nil {
		if len(matchedTarget) == 2 {
			return matchedTarget[1], nil
		}
	}
	return "", fmt.Errorf("no target found in changelogs")
}

func regexCaptureChangelogCurrentIndex(textToMatch string) (float64, error) {
	matchedCurrentIndex := changelogCurrentIndexRegexPattern.FindStringSubmatch(textToMatch)
	if matchedCurrentIndex != nil {
		if len(matchedCurrentIndex) == 2 {
			currentIndex, err := strconv.ParseFloat(matchedCurrentIndex[1], 64)
			if err != nil {
				return -1, err
			}
			return currentIndex, nil
		}
	}
	return -1, fmt.Errorf("no current index found for changelogs")
}

func regexCaptureChangelogUser(textToMatch string) [][]string {
	return changelogUserRegexPattern.FindAllStringSubmatch(textToMatch, -1)

}

type lustreLctlSource struct {
	metricCreator []lustreLctlMetricCreator
}

func newLustreLctlSource() LustreSource {
	if LctlCommandMode {
		_, err := exec.LookPath("lctl")
		if err != nil {
			log.Errorln(err)
			return nil
		}
		user, err := user.Current()
		if err != nil {
			log.Errorln(err)
			return nil
		}
		if user.Uid == "0" { // root user
			lctlCmdCall = "lctl"
			lctlCmdArgs = []string{"get_param", changelogUsersCmdArg}
		} else {
			_, err := exec.LookPath("sudo")
			if err != nil {
				log.Errorln(err)
				return nil
			}
			lctlCmdCall = "sudo"
			lctlCmdArgs = []string{"lctl", "get_param", changelogUsersCmdArg}
		}
	}
	var l lustreLctlSource
	l.metricCreator = []lustreLctlMetricCreator{}
	l.generateMDTMetricCreator(MdtEnabled)
	return &l
}

func (s *lustreLctlSource) Update(ch chan<- prometheus.Metric) (err error) {
	for _, metricCreator := range s.metricCreator {
		metricList, err := metricCreator.metricHandler(metricCreator.argument)
		if err != nil {
			return err
		}
		for _, metric := range metricList {
			ch <- metric
		}
	}
	return nil
}

func (s *lustreLctlSource) generateMDTMetricCreator(filter string) {
	if filter == extended {
		s.metricCreator = append(s.metricCreator,
			lustreLctlMetricCreator{argument: changelogUsersCmdArg, metricHandler: s.createMDTChangelogUsersMetrics})
	}
}

func (s *lustreLctlSource) createMDTChangelogUsersMetrics(text string) ([]prometheus.Metric, error) {
	metricList := make([]prometheus.Metric, 1)
	var target string
	var data string
	var err error

	if LctlCommandMode {
		out, err := exec.Command(lctlCmdCall, lctlCmdArgs...).Output()
		if err != nil {
			return nil, err
		}
		data = string(out)

	} else {
		path := strings.ReplaceAll(changelogUsersCmdArg, ".", "/")
		paths, err := filepath.Glob("lctl/" + path)
		if err != nil {
			return nil, err
		}
		if paths == nil {
			return nil, nil
		}
		out, err := ioutil.ReadFile(paths[0])
		if err != nil {
			return nil, err
		}
		data = string(out)
	}

	target, err = regexCaptureChangelogTarget(data)
	if err != nil {
		return nil, err
	}

	currentIndex, err := regexCaptureChangelogCurrentIndex(data)
	if err != nil {
		return nil, err
	}

	metricList[0] = s.counterMetric(
		[]string{"component", "target"},
		[]string{"mdt", target},
		"changelog_current_index",
		"Changelog current index.",
		currentIndex)

	// Captures registered changelog user:
	for _, changelogUserFields := range regexCaptureChangelogUser(data) {

		id := changelogUserFields[1]

		index, err := strconv.ParseFloat(changelogUserFields[2], 64)
		if err != nil {
			return nil, err
		}

		idleSeconds, err := strconv.ParseFloat(changelogUserFields[3], 64)
		if err != nil {
			return nil, err
		}

		metric := s.counterMetric(
			[]string{"component", "target", "id"},
			[]string{"mdt", target, id},
			"changelog_user_index",
			"Index of registered changelog user.",
			index)
		metricList = append(metricList, metric)

		metric = s.gaugeMetric(
			[]string{"component", "target", "id"},
			[]string{"mdt", target, id},
			"changelog_user_idle_time",
			"Idle time in seconds of registered changelog user.",
			idleSeconds)
		metricList = append(metricList, metric)
	}

	return metricList, nil
}

func (s *lustreLctlSource) counterMetric(labels []string, labelValues []string, name string, helpText string, value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.CounterValue,
		value,
		labelValues...,
	)
}

func (s *lustreLctlSource) gaugeMetric(labels []string, labelValues []string, name string, helpText string, value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.GaugeValue,
		value,
		labelValues...,
	)
}
