FROM golang:1.17.5-bullseye
MAINTAINER jknedlik <j.knedlik@gsi.de>
RUN apt-get update
COPY . /go/lustre_exporter
WORKDIR /go/lustre_exporter
#get linter and build tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0
RUN go install github.com/prometheus/promu@v0.13.0
#build lustre_exporter
RUN make
#rename /move binary
RUN mkdir /build
RUN mv lustre_exporter /build/lustre_exporter-$(cat VERSION)
#prepare mountpoint for copy
RUN mkdir /cpy
ENTRYPOINT ["/bin/bash"]
ENTRYPOINT ["cp"]
CMD ["-r","/build","/cpy/"]
