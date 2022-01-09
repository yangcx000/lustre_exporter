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
	"fmt"
	"github.com/GSI-HPC/lustre_exporter/sources"
	"gopkg.in/alecthomas/kingpin.v2"
	stdlog "log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
)

var (
	scrapeDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: sources.Namespace,
			Subsystem: "exporter",
			Name:      "scrape_duration_seconds",
			Help:      "lustre_exporter: Duration of a scrape job.",
		},
		[]string{"source", "result"},
	)
	logger = promlog.New(&promlog.Config{})
)

//LustreSource is a list of all sources that the user would like to collect.
type LustreSource struct {
	sourceList map[string]sources.LustreSource
}

//Describe implements the prometheus.Describe interface
func (l LustreSource) Describe(ch chan<- *prometheus.Desc) {
	scrapeDurations.Describe(ch)
}

//Collect implements the prometheus.Collect interface
func (l LustreSource) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(l.sourceList))
	for name, c := range l.sourceList {
		go func(name string, c sources.LustreSource) {
			collectFromSource(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
	scrapeDurations.Collect(ch)
}

func collectFromSource(name string, s sources.LustreSource, ch chan<- prometheus.Metric) {
	result := "success"
	begin := time.Now()
	err := s.Update(ch)
	duration := time.Since(begin)
	if err != nil {
		_ = level.Error(logger).Log("source %q failed after %f seconds - %s", name, duration.Seconds(), err)
		result = "error"
	} else {
		_ = level.Debug(logger).Log("source %q succeeded after %f seconds", name, duration.Seconds())
	}
	scrapeDurations.WithLabelValues(name, result).Observe(duration.Seconds())
}

func loadSources(list []string) (map[string]sources.LustreSource, error) {
	sourceList := map[string]sources.LustreSource{}
	for _, name := range list {
		fn, ok := sources.Factories[name]
		if !ok {
			return nil, fmt.Errorf("source %q not available", name)
		}
		c := fn()
		if c == nil {
			return nil, fmt.Errorf("source %q not available", name)
		}
		sourceList[name] = c
	}
	return sourceList, nil
}

func init() {
	prometheus.MustRegister(version.NewCollector("lustre_exporter"))
}

func main() {
	kingpin.Version(version.Print("lustre_exporter"))
	kingpin.HelpFlag.Short('h')

	var (
		clientEnabled       = kingpin.Flag("collector.client", "Set client metric level. Valid levels: [extended, core, disabled]").Default("extended").Enum("extended", "core", "disabled")
		genericEnabled      = kingpin.Flag("collector.generic", "Set generic metric level. Valid levels: [extended, core, disabled]").Default("extended").Enum("extended", "core", "disabled")
		lnetEnabled         = kingpin.Flag("collector.lnet", "Set LNET metric level. Valid levels: [extended, core, disabled]").Default("extended").Enum("extended", "core", "disabled")
		mdsEnabled          = kingpin.Flag("collector.mds", "Set MDS metric level. Valid levels: [extended, core, disabled]").Default("extended").Enum("extended", "core", "disabled")
		mdtEnabled          = kingpin.Flag("collector.mdt", "Set MDT metric level. Valid levels: [extended, core, disabled]").Default("extended").Enum("extended", "core", "disabled")
		mgsEnabled          = kingpin.Flag("collector.mgs", "Set MGS metric level. Valid levels: [extended, core, disabled]").Default("extended").Enum("extended", "core", "disabled")
		ostEnabled          = kingpin.Flag("collector.ost", "Set OST metric level. Valid levels: [extended, core, disabled]").Default("extended").Enum("extended", "core", "disabled")
		healthStatusEnabled = kingpin.Flag("collector.health", "Set Health metric level. Valid levels: [extended, core, disabled]").Default("extended").Enum("extended", "core", "disabled")
		listenAddress       = kingpin.Flag("web.listen-address", "Address to use to expose Lustre metrics.").Default(":9169").String()
		metricsPath         = kingpin.Flag("web.telemetry-path", "Path to use to expose Lustre metrics.").Default("/metrics").String()
		logLevel            = kingpin.Flag("log.level", "Set log level. Valid levels: [debug, info, warn, error]").Default("info").Enum("debug", "info", "warn", "error")
	)

	kingpin.Parse()

	//set Loglevel
	var allow = promlog.AllowedLevel{}
	_ = allow.Set(*logLevel)
	var config = promlog.Config{Level: &allow, Format: &promlog.AllowedFormat{}}
	logger = promlog.New(&config)

	_ = level.Info(logger).Log("Starting lustre_exporter", version.Info())
	_ = level.Info(logger).Log("Build context", version.BuildContext())

	_ = level.Info(logger).Log("Collector status:")
	sources.OstEnabled = *ostEnabled
	_ = level.Info(logger).Log(" - OST State: %s", sources.OstEnabled)
	sources.MdtEnabled = *mdtEnabled
	_ = level.Info(logger).Log(" - MDT State: %s", sources.MdtEnabled)
	sources.MgsEnabled = *mgsEnabled
	_ = level.Info(logger).Log(" - MGS State: %s", sources.MgsEnabled)
	sources.MdsEnabled = *mdsEnabled
	_ = level.Info(logger).Log(" - MDS State: %s", sources.MdsEnabled)
	sources.ClientEnabled = *clientEnabled
	_ = level.Info(logger).Log(" - Client State: %s", sources.ClientEnabled)
	sources.GenericEnabled = *genericEnabled
	_ = level.Info(logger).Log(" - Generic State: %s", sources.GenericEnabled)
	sources.LnetEnabled = *lnetEnabled
	_ = level.Info(logger).Log(" - Lnet State: %s", sources.LnetEnabled)
	sources.HealthStatusEnabled = *healthStatusEnabled
	_ = level.Info(logger).Log(" - Health State: %s", sources.HealthStatusEnabled)

	enabledSources := []string{"procfs", "procsys", "sysfs", "lctl"}

	sourceList, err := loadSources(enabledSources)
	if err != nil {
		_ = level.Error(logger).Log("Couldn't load sources: %q", err)
		os.Exit(1)
	}

	_ = level.Info(logger).Log("Available sources:")

	for s := range sourceList {
		_ = level.Info(logger).Log(" - %s", s)
	}

	prometheus.MustRegister(LustreSource{sourceList: sourceList})
	//load InstrumentMetricHandler
	handler := promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer,
		promhttp.HandlerFor(prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				ErrorLog: stdlog.New(os.Stderr, "", stdlog.LstdFlags)}))

	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var num int
		num, err = w.Write([]byte(`<html>
			<head><title>Lustre Exporter</title></head>
			<body>
			<h1>Lustre Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			_ = level.Error(logger).Log(num, err)
			os.Exit(1)
		}
	})

	_ = level.Info(logger).Log("Listening on", *listenAddress)
	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		_ = level.Error(logger).Log("Error on Listen", err)
		os.Exit(1)
	}
}
