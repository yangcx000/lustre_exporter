
# -*- docker-image-name: "xrootd_base" -*-
# xrootd base image. Provides the base image for each xrootd service

FROM golang:1.17.5-bullseye
MAINTAINER jknedlik <j.knedlik@gsi.de>
RUN apt-get update
RUN apt install vim -y
ADD lctl proc rpm sources sys systemd Gopkg.toml lustre_exporter.go lustre_exporter_test.go Makefile VERSION /go/lustre_exporter/
WORKDIR /go/lustre_exporter
RUN go mod init example.com/m/v2 
RUN go mod vendor
RUN go mod tidy -go=1.16 && go mod tidy -go=1.17
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0

# RUN go get -u github.com/golangci/golangci-lint
RUN go get -u github.com/prometheus/promu
RUN go mod vendor

ENTRYPOINT ["/bin/bash"]
#CMD ["alice-xrootd-deb/debian.deb","vol/alice-xrootd.deb"]
