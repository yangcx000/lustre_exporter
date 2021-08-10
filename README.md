# Lustre Metrics Exporter

<!-- TODO: Create an issue for both, if necessary.
[![Go Report Card](https://goreportcard.com/badge/github.com/HewlettPackard/lustre_exporter)](https://goreportcard.com/report/github.com/HewlettPackard/lustre_exporter)
[![Build Status](https://travis-ci.org/HewlettPackard/lustre_exporter.svg?branch=master)](https://travis-ci.org/HewlettPackard/lustre_exporter)
-->

[Prometheus](https://prometheus.io/) exporter for Lustre metrics.

## Getting

```
go get github.com/GSI-HPC/lustre_exporter
```
## Prerequisites

The listed versions below have been successfully used.  
Higher verions 

### Required

* [Golang](https://golang.org/)
* [Prometheus Utility Tool (promu)](https://github.com/prometheus/promu)

### Optional

* [Fast linters runner for Go (golangci-lint)](https://github.com/golangci/golangci-lint)

## Building

The build has been accomplished with the following versions successfully yet:  

* golang: 14.7
* promu: 0.7.0
* golangci-lint: 1.33.0

### Promu

Latest version:  

`go get -u github.com/prometheus/promu`

Alternatively, specific version:  

```
git clone https://github.com/prometheus/promu.git $GOPATH/src/github.com/prometheus/promu
git -C $GOPATH/src/github.com/prometheus/promu/ checkout v0.7.0
make -C $GOPATH/src/github.com/prometheus/promu/ build
cp $GOPATH/src/github.com/prometheus/promu/promu $GOPATH/bin/
```

### Golangci-lint

Latest version:  

`go get -u github.com/golangci/golangci-lint`

Alternatively, specific version:  

```
git clone https://github.com/golangci/golangci-lint.git $GOPATH/src/github.com/golangci/golangci-lint/
git -C $GOPATH/src/github.com/golangci/golangci-lint/ checkout v1.33.0
make -C $GOPATH/src/github.com/golangci/golangci-lint/ build
cp $GOPATH/src/github.com/golangci/golangci-lint/golangci-lint $GOPATH/bin/
```

### Exporter

For just building the exporter:

```
cd $GOPATH/src/github.com/GSI-HPC/lustre_exporter
make build
```

Building the exporter with code testing, formatting and linting:

```
cd $GOPATH/src/github.com/GSI-HPC/lustre_exporter
make
```

### RPM Package Build

[Manual RPM Package Creation](rpm/README.md)

## Running

```
./lustre_exporter <flags>
```

### Flags

* collector.ost=disabled/core/extended
* collector.mdt=disabled/core/extended
* collector.mgs=disabled/core/extended
* collector.mds=disabled/core/extended
* collector.client=disabled/core/extended
* collector.generic=disabled/core/extended
* collector.lnet=disabled/core/extended
* collector.health=disabled/core/extended

All above flags default to the value "extended" when no argument is submitted by the user.

Example: `./lustre_exporter --collector.ost=disabled --collector.mdt=core --collector.mgs=extended`

The above example will result in a running instance of the Lustre Exporter with the following statuses:
* collector.ost=disabled
* collector.mdt=core
* collector.mgs=extended
* collector.mds=extended
* collector.client=extended
* collector.generic=extended
* collector.lnet=extended
* collector.health=extended

Flag Option Detailed Description

- disabled - Completely disable all metrics for this portion of a source.
- core - Enable this source, but only for metrics considered to be particularly useful.
- extended - Enable this source and include all metrics that the Lustre Exporter is aware of within it.

## What's exported?

All Lustre procfs and procsys data from all nodes running the Lustre Exporter that we perceive as valuable data is exported or can be added to be exported (we don't have any known major gaps that anyone cares about, so if you see something missing, please file an issue!).

See the issues tab for all known issues.

## Troubleshooting

In the event that you encounter issues with specific metrics (especially on versions of Lustre older than 2.7), please try disabling those specific troublesome metrics using the documented collector flags in the 'disabled' or 'core' state. Users have encountered bugs within Lustre where specific sysfs and procfs files miscommunicate their sizes, causing read calls to fail.

## Contributing

You are welcome to contribute to the project.
Feel free to create an issue, pull request or just start a discussion.
