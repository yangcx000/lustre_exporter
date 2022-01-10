
# -*- docker-image-name: "xrootd_base" -*-
# xrootd base image. Provides the base image for each xrootd service

FROM golang:1.17.5-bullseye
MAINTAINER jknedlik <j.knedlik@gsi.de>
RUN apt-get update
COPY . /go/lustre_exporter
WORKDIR /go/lustre_exporter
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN go get -u github.com/prometheus/promu
ENTRYPOINT ["/bin/bash"]
