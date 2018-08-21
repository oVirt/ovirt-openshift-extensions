# ovirt-openshift-extensions

[![Build Status](http://jenkins.ovirt.org/buildStatus/icon?job=oVirt_ovirt-openshift-extensions_standard-on-ghpush)](http://jenkins.ovirt.org/job/oVirt_ovirt-openshift-extensions_standard-on-ghpush/)
[![Go Report Card](https://goreportcard.com/badge/github.com/ovirt/ovirt-openshift-extensions)](https://goreportcard.com/report/github.com/ovirt/ovirt-openshift-extensions)

## Purpose
The project purpose is to the best out of openshift installation on top of oVirt.
It's main components are:

### ovirt-volume-provisioner
A kubernetes controller that creates/deletes persistent volumes, and allocates disks \
in ovirt as a result. This is the first part for providing volumes from ovirt.

### ovirt-flexvolume-driver
A kubernetes node plugin that attaches/detaches a volume to a container. \
It attaches the ovirt disk to the kube node(which is an ovirt vm). It identifies the disk device on the os, \
prepares a filesystem, then mounts it so it is ready as a volume mount into a container.

### ovirt-cloud-provider
An out-of-tree implementation of a cloudprovider. \
A controller that manages the admission of new nodes for openshift, from ovirt vms. \
Merging this code is work in progress here: https://github.com/oVirt/ovirt-openshift-extensions/pull/59 \
NOTE: ovirt-cloud-provider will be available in v0.3.3

### Versions

\<= v0.3.1 - oVirt >= 4.2, Openshift origin 3.9, OKD 3.10 \
\>= v0.3.2 - oVirt >= 4.2, OKD 3.10, OKD 3.11 \

### Installation
There are 2 main deployment methods: using a deployment container(recommended) or manual

1. Deploy using the deployment container(APB) and service-catalog(recommended)

Pre-requisite:
- Openshift 3.10.0
- Running service catalog

From the repo:
- push the apb image to your cluster repo
   ```
   make apb_build apb_push
   ```
- go to the service catalog UI and deploy the ovirt-flexvolume-driver-apb. \
 Here is a demo doing that: \
<a href="http://www.youtube.com/watch?feature=player_embedded&v=frcehKUk_g4" target="_blank"><img src="http://img.youtube.com/vi/frcehKUk_g4/0.jpg" alt="IMAGE ALT TEXT HERE" width="240" height="180" border="10" /></a>

2. Deploy Manually

- make sure `oc` command is configured and has access to your cluster, e.g run `oc status`

- use a cluster admin user to deploy, or grant permission to one
   ```
   oc login -u system:admin
   oc adm policy add-cluster-role-to-user cluster-admin developer
   ```
- Run on a master:
   ```
   OCP_USER=developer \
   OCP_PASS=pass \
   ENGINE_URL=https://engine-fqdn/ovirt-engine/api \
   ENGINE_USER=admin@internal \
   ENGINE_PASS=123
   
   docker run \
    --rm \
    --net=host \
    -v $HOME/.kube:/opt/apb/.kube:z \
    -u $UID docker.io/rgolangh/ovirt-flexvolume-driver-apb \
    provision \
    -e admin_password=$OCP_PASS -e admin_user=$OCP_USER \
    -e cluster=openshift -e namespace=default \
    -e engine_password=$ENGINE_PASS -e engine_url=$ENGINE_URL \
    -e engine_username=$ENGINE_USER
   ```

   - If its the first time deploying the image then it should take few moments to download it.

Upon completion you have the components running:

   ```
   oc get ds -n default ovirt-flexvolume-driver 
   name                      desired   current   ready     up-to-date   available   node selector   age
   ovirt-flexvolume-driver   1         1         1         1            1           <none>          15m

   oc get deployment -n default ovirt-volume-provisioner 
   NAME                       DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
   ovirt-volume-provisioner   1         1         1            1           17m
   ```

### Contributing
**Feedback is most welcome**, if you have and idea, proposal, fix, or want to chat \
  you'll find the details here: 
- Patches are usual pull-request, see the pull request template.
- Upstream bugs: https://github.com/oVirt/ovirt-openshift-extensions/issues
- Trello Board: ovirt-openshift-extensions (may change to Jira)
- Bugzilla tracker bug: https://bugzilla.redhat.com/show_bug.cgi?id=1581638 (see the bugzilla tracker for the component under rhev)
- IRC: upstream: #ovirt @ oftc.net
- Mailing List: devel@ovirt.org

Blog post in ovirt: https://www.ovirt.org/blog/2018/02/your-container-volumes-served-by-ovirt/

Youtube Demo: https://youtu.be/_E9pUVrI0hs


### References
- [flexvolume plugin page on openshift](https://docs.openshift.org/latest/install_config/persistent_storage/persistent_storage_flex_volume.html)
- [flexvolume spec on kubernetes page](https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md)

[flex-conf]: deployment/ovirt-flexdriver/ovirt-flexdriver.conf.j2
[flex-playbook]: deployment/ovirt-flexdriver/deploy.yaml
[prov-playbook]: deployment/ovirt-provisioner/deploy.yaml
