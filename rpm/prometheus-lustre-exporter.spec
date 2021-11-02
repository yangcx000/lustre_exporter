%define        __spec_install_post %{nil}
%define          debug_package %{nil}
%define        __os_install_post %{_dbpath}/brp-compress

Name:           prometheus-lustre-exporter
Version:        2.1.1
Release:        1%{?dist}
Summary:        Prometheus exporter for use with the Lustre parallel filesystem
Group:          Monitoring

License:        ASL 2.0
URL:            https://github.com/GSI-HPC/lustre_exporter
Source0:        %{name}-%{version}.tar.gz

Requires(pre): shadow-utils

Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd
%{?systemd_requires}
BuildRequires:  systemd

BuildRoot:      %{_tmppath}/%{name}-%{version}-1-root

%description
The Lustre exporter for Prometheus will expose all Lustre procfs and procsys data.

%prep
%setup -q

%build
# Empty section.

%install
rm -rf %{buildroot}
mkdir -p  %{buildroot}
mkdir -p %{buildroot}%{_unitdir}/
cp usr/lib/systemd/system/%{name}.service %{buildroot}%{_unitdir}/

# in builddir
cp -a * %{buildroot}

%clean
rm -rf %{buildroot}

%pre
getent group prometheus >/dev/null || groupadd -r prometheus
getent passwd prometheus >/dev/null || \
    useradd -r -g prometheus -d /dev/null -s /sbin/nologin \
    -c "Prometheus exporter user" prometheus
mkdir -p /etc/sudoers.d/
echo "prometheus ALL = NOPASSWD: /usr/sbin/lctl get_param *" > /etc/sudoers.d/prometheus
chmod 0440 /etc/sudoers.d/prometheus
exit 0

%post
systemctl enable %{name}.service
systemctl start %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun_with_restart %{name}.service

%files
%defattr(-,root,root,-)
%config /etc/sysconfig/prometheus-lustre-exporter.options
%{_bindir}/lustre_exporter
%{_unitdir}/%{name}.service
