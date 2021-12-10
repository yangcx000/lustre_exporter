# Manual RPM Package Creation

Given the Lustre exporter version 2.1.2.

## Create Directory Structure for RPM Build

Top level directory structure:

`mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}`  

Exporter directory:  

`mkdir -p ~/rpmbuild/SOURCES/prometheus-lustre-exporter-2.1.2/usr/bin/`  
`mkdir -p ~/rpmbuild/SOURCES/prometheus-lustre-exporter-2.1.2/usr/lib/systemd/system/`  
`mkdir -p ~/rpmbuild/SOURCES/prometheus-lustre-exporter-2.1.2/etc/sysconfig/`  

Copy required files from exporter source directory to RPM exporter directory:  

`cp $GOPATH/src/github.com/GSI-HPC/lustre_exporter/rpm/prometheus-lustre-exporter.spec ~/rpmbuild/SPECS/`  
`cp $GOPATH/src/github.com/GSI-HPC/lustre_exporter/lustre_exporter ~/rpmbuild/SOURCES/prometheus-lustre-exporter-2.1.2/usr/bin/`  
`cp $GOPATH/src/github.com/GSI-HPC/lustre_exporter/systemd/prometheus-lustre-exporter.service ~/rpmbuild/SOURCES/prometheus-lustre-exporter-2.1.2/usr/lib/systemd/system/`  
`cp $GOPATH/src/github.com/GSI-HPC/lustre_exporter/systemd/prometheus-lustre-exporter.options ~/rpmbuild/SOURCES/prometheus-lustre-exporter-2.1.2/etc/sysconfig/`  

After setup the RPM top level directory should contain the required files:  

`tree ~/rpmbuild/`  
```
├── BUILD
├── RPMS
├── SOURCES
│   └── prometheus-lustre-exporter-2.1.2
│       ├── etc
│       │   └── sysconfig
│       │       └── prometheus-lustre-exporter.options
│       └── usr
│           ├── bin
│           │   └── lustre_exporter
│           └── lib
│               └── systemd
│                   └── system
│                       └── prometheus-lustre-exporter.service
├── SPECS
│   └── prometheus-lustre-exporter.spec
└── SRPMS
```
Create source TAR ball for RPM build:  

`cd ~/rpmbuild/SOURCES; tar -czvf prometheus-lustre-exporter-2.1.2.tar.gz prometheus-lustre-exporter-2.1.2`  

> Use relative path here, otherwise rpmbuild will not find the extracted files.  

Build the RPM package:  

`rpmbuild -ba ~/rpmbuild/SPECS/prometheus-lustre-exporter.spec`  
