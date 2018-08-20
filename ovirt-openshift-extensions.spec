Name:       ovirt-openshift-extensions
Version:    %{?_version}
Release:    %{?_release}%{?dist}
License:    ASL 2.0
URL:        http://www.ovirt.org
Source0:    %{name}-%{version}%{?_release:-%_release}.tar.gz
Summary:    flexvolume and provisioner
ExclusiveArch: x86_64

%description
A Flexvolume driver to provision k8s volumes using oVirt

%package -n ovirt-flexvolume-driver
Summary:    Flexvolume driver k8s using oVirt

%description -n ovirt-flexvolume-driver
A Flexvolume driver to provision k8s volumes using oVirt

%package -n ovirt-provisioner
Summary:    Storage provisioner plugin for k8s using oVirt

%description -n ovirt-provisioner
Storage provisioner plugin for k8s using oVirt

%package -n ovirt-cloud-provider
Summary:   Cloud provider controller for oVirt 

%description -n ovirt-cloud-provider
Cloud provider controller for oVirt

%global vendor ovirt
%global kube_plugin_dir   /usr/libexec/kubernetes/kubelet-plugins/volume/exec/%{vendor}~ovirt-flexvolume-driver
%global org github.com/ovirt
%global repo ovirt-openshift-extensions
%global golang_version 1.9.1
%global debug_package %{nil}

BuildRequires: golang >= %{golang_version}

%prep
%setup -c -q

%build
# set up temporary build gopath for the rpmbuild
set -x
pwd

%define tmp_go_path build
mkdir -p ./%{tmp_go_path}/src/%{org}
ln -s $(pwd) ./build/src/%{org}/%{repo}

export GOPATH=$(pwd)/%{tmp_go_path}
cd %{tmp_go_path}/src/%{org}/%{repo}
make build

%install
mkdir -p %{buildroot}/%{kube_plugin_dir}
install -p -m 755 ovirt-flexdriver %{buildroot}/%{kube_plugin_dir}/ovirt-flexvolume-driver
install -p -m 644 deployment/ovirt-flexdriver/ovirt-flexdriver.conf.j2 %{buildroot}//%{kube_plugin_dir}/ovirt-flexvolume-driver.conf
mkdir -p %{buildroot}/usr/bin/
install -p -m 755 ovirt-provisioner %{buildroot}/usr/bin/
install -p -m 755 ovirt-cloud-provider %{buildroot}/usr/bin/

%files -n ovirt-flexvolume-driver
%dir %{kube_plugin_dir}
%{kube_plugin_dir}/ovirt-flexvolume-driver
%{kube_plugin_dir}/ovirt-flexvolume-driver.conf

%files -n ovirt-provisioner
/usr/bin/ovirt-provisioner

%files -n ovirt-cloud-provider
/usr/bin/ovirt-cloud-provider

