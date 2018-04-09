# ovirt-flexdriver

[![Go Report Card](https://goreportcard.com/badge/github.com/rgolangh/ovirt-flexdriver)](https://goreportcard.com/report/github.com/rgolangh/ovirt-flexdriver)

Implementation of flexvolume driver for [oVirt](https://ovirt.org) and a dynamic volume provisioner

oVirt flexvolume driver is attachable, i.e. it supports attaching/detaching storage volumes from nodes, by detaching them from the underlying VM.

Here is a short [demo](http://www.youtube.com/watch?v=_E9pUVrI0hs):\
<a href="http://www.youtube.com/watch?feature=player_embedded&v=_E9pUVrI0hs" target="_blank"><img src="http://img.youtube.com/vi/_E9pUVrI0hs/0.jpg" 
alt="IMAGE ALT TEXT HERE" width="240" height="180" border="10" /></a>

The project creates 3 containers:
1. **`ovirt-flexvolume-driver`**
   A container that exist to expose the binary, imediatly sleeps
forever. Used in a daemonset deployment - see the apb.
2. **`ovirt-volume-provisioner`**
   A container for the provisioer controller. Used in a deployment - see
apb
3. **`ovirt-flexvolume-driver-apb`**
  An apb container that will deploy both the driver and the provisioner.
  One can use the service catalog to push the apb there and use it or straight from the command line.
  See the apb Makefile currently under [deployment/ovirt-flexvolume-driver/Makefile](deployment/ovirt-flexvolume-driver/Makefile).

# Deployment
There are 2 main deployment methods: using a deployment container(recommended) or manual

## Deploy using the deployment container(APB) and service-catalog(recommended)

Pre-requisite:
- Openshift 3.9.0
- Running service catalog

1. push the apb image to your cluster repo
   ```
   cd deployment/ovirt-flexvolume-driver
   make apb_push DOCKERHOST=your_registry
   ```
2. go to the service catalog UI and deploy the ovirt-flexvolume-driver-apb. Here is a demo doing that:
<a href="http://www.youtube.com/watch?feature=player_embedded&v=frcehKUk_g4" target="_blank"><img src="http://img.youtube.com/vi/frcehKUk_g4/0.jpg" alt="IMAGE ALT TEXT HERE" width="240" height="180" border="10" /></a>

## Deploy Manually

1. push the apb image like in the service-catalog deployment

2. use docker directly to invoke the deployment
   - replace 172.30.1.1:5000/openshift/ovirt-flexvolume-driver-apb with your image url
   - fill in the details in `extra-vars`
   
   ```
   docker run \
     --rm \
     --net=host \
     -v $HOME/.kube:/opt/apb/.kube:z \
     -u $UID 172.30.1.1:5000/openshift/ovirt-flexvolume-driver-apb \
     provision \
     --extra-vars '{"admin_password":"developer","admin_user":"developer","cluster":"openshift","namespace":"default","engine_password":"123","engine_url":"https://your_engine_hostname:28443/ovirt-engine/api","engine_username":"admin@internal"}'
   ```

# Pre-requisite
  - Running ovirt 4.1 instance (support for disk_attachments API)
  - k8s 1.9 (possibly working on 1.8, untested)
  - Every k8s minion name should match its VM name


# Make targets
There are few make targets for building the artifacts:
- `make deps`      - get and install the project dependencies
- `make build`     - build the flexvolume driver and provisioner binaries
- `make rpm`       - builds and rpm from the previously created binaries
- `make container` - creates the containers

# Sources
- [flexvolume plugin page on openshift](https://docs.openshift.org/latest/install_config/persistent_storage/persistent_storage_flex_volume.html)
- [flexvolume spec on kubernetes page](https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md)

[flex-conf]: deployment/ovirt-flexdriver/ovirt-flexdriver.conf.j2
[flex-playbook]: deployment/ovirt-flexdriver/deploy.yaml
[prov-playbook]: deployment/ovirt-provisioner/deploy.yaml
