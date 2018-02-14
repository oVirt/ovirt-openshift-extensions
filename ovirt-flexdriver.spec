Name:       ovirt-flexdriver
Version:    %{?_version}
Release:    %{?_release}
Summary:    A Flexvolume driver to provision k8s volumes using oVirt

License:    ASL 2.0
URL:        http://www.ovirt.org
Source0:    %{name}-%{version}%{?release:-%release}.tar.gz

%description
A Flexvolume driver to provision k8s volumes using oVirt

%global vendor ovirt
%global kube_plugin_dir   /usr/libexec/kubernetes/kubelet-plugins/volume/exec/%{vendor}~%{name}

%prep
%setup -c

%build
make deps
make build-flex

%install
mkdir -p %{buildroot}%{kube_plugin_dir}

install -p -m 755 %{name} %{buildroot}%{kube_plugin_dir}
install -p -m 644 deployment/%{name}/%{name}.conf.j2 %{buildroot}%{kube_plugin_dir}/%{name}.conf

%define debug_package %{nil}

%files
%dir %{kube_plugin_dir}
%{kube_plugin_dir}/%{name}
%{kube_plugin_dir}/%{name}.conf

%changelog
