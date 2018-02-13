# ovirt-flexdriver

[![Go Report Card](https://goreportcard.com/badge/github.com/rgolangh/ovirt-flexdriver)](https://goreportcard.com/report/github.com/rgolangh/ovirt-flexdriver)

Implementation of flexvolume driver for [oVirt](https://ovirt.org) and a dynamic volume provisioner


oVirt flexvolume driver is attachable, i.e. it supports attaching/detaching storage volumes from nodes, by detaching them from the underlying VM.


# Deployment

## Deploy the Flexvolume driver
The driver is a binary executable that needs to reside on every node,which is an oVirt VM,
and every master.\
The [deploy.yaml][flex-playbook] playbook
resolves configuration template [ovirt-flexdriver.conf.j2][flex-conf]  
- fill in the details of your environment ``
  ```ini
  [general]
  # the inventory needs to map each hostname to the vm name in oVirt
  ovirtVmName={{ vm_name }}
  
  [connection]
  url=https://hostname/ovirt-engine/api
  username=admin@internal
  password=pass
  insecure=false
  cafile=
  ```
- [Build](#Build) the binary or download from the release\
For that use the [deploy.yaml][flex-playbook] Ansible playbook.\
Your inventory needs to specify 2 groups, one for nodes and one for masters:
  ```ini
  # ansible inventory
  ...
  [k8s-ovirt-nodes]
  kube-node1 vm_name=kube-node1
  kube-node2 vm_name=kube-node2
  [k8s-ovirt-masters]
  kube-master1 vm_name=kube_master
   ```

- run the playbook to deploy the flexdriver
  ```
  ansible-playbook deployment/ovirt-flexdriver/deploy.yaml
  ```

- Pre-requisite
  - Running ovirt 4.1 instance (support for disk_attachments API)
  - k8s 1.9 (possibly working on 1.8, untested)
  - Every k8s minion name should match its VM name

## Deploy the dynamic provisioner POD
The provisioner is a container which registers itself to the [kubernetes provision controller]()\
so all we need is to create the needed definitions in k8s and let it start it. The definitions are\
all under [ovirt-provisioner-manifest.yaml](deployment/ovirt-provisioner/k8s/ovirt-provisioner-manifest.yaml.j2).

Now edit the vars section in the [playbook][prov-playbook] and run it:
```
ansible-playbook deployment/ovirt-provisioner/deploy.yaml
```
  
# Build
There are few make targets for building the artifacts:
- `make build-flex` - build the flexdriver in local 
- `make build-provisioner` - build the provisioner
- `make build` - build the flexvolume driver and provisioner
- `make quick-container` - creates the container for the provisioner, with a tag derived from `git describe`
  
# Sources
- [flexvolume plugin page on openshift](https://docs.openshift.org/latest/install_config/persistent_storage/persistent_storage_flex_volume.html)
- [flexvolume spec on kubernetes page](https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md)

[flex-conf]: deployment/ovirt-flexdriver/ovirt-flexdriver.conf.j2
[flex-playbook]: deployment/ovirt-flexdriver/deploy.yaml
[prov-playbook]: deployment/ovirt-provisioner/deploy.yaml