# ovirt-flexdriver

[![Go Report Card](https://goreportcard.com/badge/github.com/rgolangh/ovirt-flexdriver)](https://goreportcard.com/report/github.com/rgolangh/ovirt-flexdriver)

Implementation of flexvolume driver for [oVirt](https://ovirt.org) and a dynamic volume provisioner

oVirt flexvolume driver is attachable, i.e. it supports attaching/detaching storage volumes from nodes, by detaching them from the underlying VM.

Here is a short [demo](http://www.youtube.com/watch?v=_E9pUVrI0hs):\
<a href="http://www.youtube.com/watch?feature=player_embedded&v=_E9pUVrI0hs" target="_blank"><img src="http://img.youtube.com/vi/_E9pUVrI0hs/0.jpg" 
alt="IMAGE ALT TEXT HERE" width="240" height="180" border="10" /></a>

# Deployment
Both ovirt-flexdriver and ovirt-provisioner have deployment containers that use Ansible -\
`ovirt-flexdriver-ansible` and `ovirt-provisioner-ansible`.

* Prepare an inventory file
fill in the details of your nodes and master, and the ovirt-engine api connection details:
 
    ```ini
    # ansible inventory - either /etc/ansible/hosts or a custom one
    
    [k8s-ovirt-nodes]
    kube-node1 vm_name=kube-node1
    kube-node2 vm_name=kube-node2
    
    [k8s-ovirt-masters]
    kube-master1 vm_name=kube_master
    
    [all:vars]
    engine_url=https://hostname/ovirt-engine/api
    engine_username=admin@internal
    engine_password=123
    engine_insecure=false
    engine_ca_file=
    ```

* run the Ansible containers to deploy
    ```
    sudo docker run --rm -it \
        -v /root/.ssh:/root/.ssh:z \
        -v /etc/ansible/hosts:/etc/ansible/hosts \
        rgolangh/ovirt-flexdriver-ansible:v0.2.0
    
    sudo docker run --rm -it \
        -v /root/.ssh:/root/.ssh:z \
        -v /etc/ansible/hosts:/etc/ansible/hosts \
        rgolangh/ovirt-flexdriver-ansible:v0.2.0
    ```

- Pre-requisite
  - Running ovirt 4.1 instance (support for disk_attachments API)
  - k8s 1.9 (possibly working on 1.8, untested)
  - Every k8s minion name should match its VM name

  
# Build in a container
Building is done in dedicated containers, or can be done manually:
Make targets for containers:
- `make container-flexdriver`  - build the flexdriver in a container with the Ansible deploy playbook
- `make container-provisioner-binary`  - build the provisioner container
- `make container-provisioner-ansible` - build the provisioner Ansible container with the deploy playbook
- `make container` - build all containers

# Build locally
There are few make targets for building the artifacts:
- `make deps` - get and install the project dependencies
- `make build-flex` - build the flexdriver in local 
- `make build-provisioner` - build the provisioner
- `make build` - build the flexvolume driver and provisioner
- `make quick-container` - creates the container for the provisioner, with a tag derived from `git describe`
- `make rpm` - builds and rpm from the previously created binaries

# Sources
- [flexvolume plugin page on openshift](https://docs.openshift.org/latest/install_config/persistent_storage/persistent_storage_flex_volume.html)
- [flexvolume spec on kubernetes page](https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md)

[flex-conf]: deployment/ovirt-flexdriver/ovirt-flexdriver.conf.j2
[flex-playbook]: deployment/ovirt-flexdriver/deploy.yaml
[prov-playbook]: deployment/ovirt-provisioner/deploy.yaml
