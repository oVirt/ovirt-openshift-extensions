# ovirt-flexdriver

[![Go Report Card](https://goreportcard.com/badge/github.com/rgolangh/ovirt-flexdriver)](https://goreportcard.com/report/github.com/rgolangh/ovirt-flexdriver)

Implementation of flexvolume driver for [oVirt](https://ovirt.org).

This is WIP 

oVirt flexvolume driver is attachable, i.e. it supports attaching/detaching storage volumes from nodes, by detaching them from the underlying VM.

# Sources
- [flexvolume plugin page on openshift](https://docs.openshift.org/latest/install_config/persistent_storage/persistent_storage_flex_volume.html)
- [flexvolume spec on kubernetes page](https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md)

# Deployment
This driver impl is still in its early days so the deployment is nothing fancy, just copying the driver binary and config into each master/node in k8s
For that use the [deployment/deploy](https://github.com/rgolangh/ovirt-flexdriver/blob/master/deployment/deploy.yaml) Ansible playbook.
Your inventory needs to hold 2 groups, one for nodes and one for masters:
```ini
[k8s-ovirt-nodes]
kube-node1
kube-node1
[k8s-ovirt-masters]
kube-master1
kube-master2
```

- Pre-requisite
  - Running ovirt 4.1 instance (support for disk_attachments API)
  - Every k8s minion name should match its VM name