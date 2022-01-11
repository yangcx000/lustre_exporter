# Lustre Metrics Exporter

<!-- TODO: Create an issue for both, if necessary.
[![Go Report Card](https://goreportcard.com/badge/github.com/HewlettPackard/lustre_exporter)](https://goreportcard.com/report/github.com/HewlettPackard/lustre_exporter)
[![Build Status](https://travis-ci.org/HewlettPackard/lustre_exporter.svg?branch=master)](https://travis-ci.org/HewlettPackard/lustre_exporter)
-->

[Prometheus](https://prometheus.io/) exporter for Lustre metrics.

## Getting

```
go install github.com/GSI-HPC/lustre_exporter@latest
```
## Prerequisites

The listed versions below have been successfully used.  

### Required

* [Golang](https://golang.org/)
* [Prometheus Utility Tool (promu)](https://github.com/prometheus/promu)

### Optional

* [Fast linters runner for Go (golangci-lint)](https://github.com/golangci/golangci-lint)

## Building

The build has been accomplished with the following versions successfully yet:  

* golang: 1.17.5
* promu: 0.13.0
* golangci-lint: 1.43.0


### Promu

Use Promu v0.13.0 version:  

`go install github.com/prometheus/promu@v0.13.0`


### Golangci-lint

Use golangci-lint v1.43.0
Latest version:  

`go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0`


### Exporter

For just building the exporter: (in the repo base dir)

```
make build
```

Building the exporter with code testing, formatting and linting:

```
make
```

### RPM Package Build

### Build Containers

Two Docker container for building the Lustre Exporter are provided:

1.  A simple container based on the offical golang:1.17.5-bullseye container image that provides the `lustre_exporter` binary. 
Build it via 
```shell
# from repo base dir run
docker build  --tag l_export -f build_containers/Dockerfile .
docker run -v $PWD:/cpy -it lustre_exporter
```
Your binary is then available in build/lustre_exporter-X.X.X

2. A Centos7 container based on the official Centos7 container image that provides a rpm package containing the Lustre Exporter and systemd unit files
Build it via 
```shell
# from repo base dir run
docker build -t rpm_dock -f build_containers/RPM-Dockerfile .
docker run -v $PWD:/rpm -it rpm_dock
```
Your rpm package is then available in `build/prometheus-lustre-exporter-vX.X.X-X.X.el7.x86_64.rpm`

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
