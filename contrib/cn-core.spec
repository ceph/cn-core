%global source_version 0.6
%global tag 1
%global provider        github
%global provider_tld    com
%global gopath          %{_datadir}/gocode

Name:           cn
%global project         %{name}
%global repo            %{name}
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path     %{provider_prefix}
Version:        %{source_version}
Release:        %{tag}%{?dist}
Summary:        A binary to deploy a Ceph AIO, used by cn (ceph-nano)
License:        Apache-2.0
Group:          System/Filesystems
URL:            https://%{import_path}
Source0:        https://%{import_path}/archive/v%{source_version}.tar.gz
Source1:        %{name}-vendor-%{source_version}.tar.xz
Source2:        rebuild-vendor.sh

%if !%{defined gobuild}
%define gobuild(o:) go build -compiler gc -ldflags "${LDFLAGS:-} -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n')" -a -v -x %{?**};
%endif

BuildRequires:  go-srpm-macros
BuildRequires:  golang
BuildRequires:  dep

%description
A binary to deploy a Ceph AIO, used by cn (ceph-nano).

%prep
%setup -q -a 1 -n cn-%version

# move content of vendor under Godeps
mkdir -p Godeps/_workspace/src
mv vendor/* Godeps/_workspace/src/

%build
export GOPATH=$(pwd):$(pwd)/Godeps/_workspace:%{gopath}
export LDFLAGS="$LDFLAGS -X main.version=%{source_version}"
%gobuild -o bin/cn main.go

%install
install -D -p -m 755 bin/cn-core %{buildroot}%{_bindir}/cn
install -D -p -m 644 cn-core.toml %{buildroot}%{_sysconfdir}/cn/

%files
%doc README.md
%{_bindir}/cn-core
%{_sysconfdir}/cn/cn-core.toml

%changelog
* Wed Apr 10 2019  Sebastien Han <seb@redhat.com> - 0.6-1
- config: shrink ceph.conf
- contrib: adds latest tag
- dockerfile: updates base image
- rgw: change default frontend to beast
- Packaging: Update specfile version to v0.5
* Wed Mar 13 2019  Sebastien Han <seb@redhat.com> - 0.5-1
- utils: fix dashboard configuration
- daemon startup CLI: fix flags
- Packaging: Update specfile version to v0.4
* Thu Feb 14 2019  Sebastien Han <seb@redhat.com> - 0.4-1
- mon: updates mon host address
- Adds ceph-volume support
- travis: golint
- Packaging: Update specfile version to v0.3
* Tue Jan 22 2019  Sebastien Han <seb@redhat.com> - 0.3-1
- Merge pull request #7 from ceph/memory-tuning
- osd: implement memory limit
- main: use log.Fatal
- Merge pull request #9 from ceph/mon-msgr2
- mon: enable msgr2
- Merge pull request #5 from ceph/improve-key-gen
- mon: simplify key generation
- travis: remove leftovers
- Packaging: Update specfile version to v0.2
* Wed Jan 16 2019  Sebastien Han <seb@redhat.com> - 0.2-1
- contrib: fix release building
# nothing yet
