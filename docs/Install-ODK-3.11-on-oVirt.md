# Install OKD 3.11 on oVirt

> DRAFT - WIP

Installing OKD on oVirt has many advantages, and it's also got a lot easier these days. Admins and users who like to take container platform management for a spin on oVirt will be encouraged by this.

The installation uses the [openshift-ansible](https://github.com/openshift/openshift-ansible) and specifically the `openshift_ovirt` ansible role. The integration between openshift and oVirt is tighter, and provides storage integration. If you need persistent volumes for your containers, you can get that directly from oVirt using the **ovirt-volume-provisioner** and **ovirt-flexvolume-driver**.

For the sake of simplicity, this installation will be an all-in-one Openshift cluster, on a single VM. On top of that
we would run a classic web stack - Node.js + Postgres. Postgres will get a persistent volume from oVirt using its flexvolume driver.

## Single shell file installation

Dropping to shell - this [install.sh](https://github.com/oVirt/ovirt-openshift-extensions/blob/master/automation/ci/install.sh) is a wrapper for installation of ovirt-openshift-installer container, it uses asible-playbook and has 2 main playbooks - 1 - install_okd.yaml and install_extensions.yaml. The latter is mainly for installing oVirt storage plugins.

The install.sh has one dependency - it needs to have 'podman' installed. All the rest runs inside a container.
`dnf install podman` will do, [for other ways to install podman consult the readme](https://github.com/containers/libpod/blob/master/docs/tutorials/podman_tutorial.md).

### Get the install.sh and customize
```bash
curl -O "https://raw.githubusercontent.com/oVirt/ovirt-openshift-extensions/master/automation/ci/{install.sh,vars.yaml}"
```

Edit the `vars.yaml`:

- Put the engine details in engine_url
  ```yaml
  engine_url: https://ovirt-engine-fqdn/ovirt-engine/api
  ```

- Choose the oVirt cluster and data domain unless you want to use 'Default'
  ```yaml
  openshift_ovirt_cluster: yours
  openshift_ovirt_data_store: yours
   ```

- Unmark the memory check if you don't change the VM default memory (for this installation it's 8Gb)
  ```yaml
  openshift_disable_check: memory_availability,disk_availability,docker_image_availability
  ```

- Domain name of the setup. The setup will create a VM with the name master0.$public_hosted_zone here. It will
  be used for all the components of the setup
  ```yaml
  public_hosted_zone: example.com
  ```

For a more complete list of customization please [refer to the vars.yaml](https://github.com/oVirt/ovirt-openshift-extensions/blob/master/automation/ci/vars.yaml) packed into the container.

## Install

Run the install.sh to start the installation.

```console
[user@bastion ~]# bash install.sh
```

What it would do is the following:
1. Pull the ovirt-openshift-installer container and run it
2. Download CentOS cloud image and import it into oVirt (set by `qcow_url`)
3. Create a VM named master0.example.com based on the template above (domain name is set by `public_hosted_zone`)
4. Cloud-init will configure repositories, network, ovirt-guest-agent, etc. (set by `cloud_init_script_master`)
5. The VM will dynamically be inserted into an ansible inventory, under `master`, `compute`, and `etc` groups
6. Openshift-ansible main playbooks are executed - `prerequisite.yml` and `deploy_cluster.yml` to install OKD

In the end there is an all-in-one cluster running. Let's check it out.


```console
[root@master0 ~]# oc get nodes
NAME                         STATUS    ROLES                  AGE       VERSION
master0.example.com   Ready     compute,infra,master   1h        v1.11.0+d4cacc0
```

Check oVirt's extensions
```console
[root@master0 ~]# oc get deploy/ovirt-volume-provisioner
NAME                       DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
ovirt-volume-provisioner   1         1         1            1           57m

[root@master0 ~]# oc get ds/ovirt-flexvolume-driver
NAME                      DESIRED   CURRENT   READY     UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
ovirt-flexvolume-driver   1         1         1         1            1           <none>          59m
```

In case the router is not scheduled, label the node with this:
```console
[root@master0 ~]# oc label node master0.example.com  "node-router=true"
```

To make all the dynamic storage provisioning run through oVirt's provisioner, \
we need to set oVirt's storage class the default. Note a storage class defines which oVirt storage domain will \
be used to provision the disks. Also it will set the disk type (thin/thick provision) with thin being the default.

```console
[root@master ~]# oc patch sc/ovirt -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

[![asciicast](https://asciinema.org/a/219956.svg)](https://asciinema.org/a/219956)

Ready to go! Let's deploy mini messaging app written in java + postgresql with a persistent disk from oVirt.

### Deploy persistent Postgres container with its storage from oVirt

A persistent deployment means that `/var/lib/pgsql/data` where the data is saved will be kept on a persistent
volume disk. First let's pull a template of a persistent postgres deployment:

```console
[root@master ~]# curl -O https://raw.githubusercontent.com/openshift/library/master/arch/x86_64/official/postgresql/templates/postgresql-persistent.json
```

Create a new-app based on this deployment. The parameters will create the proper persistent volume claim:

```console
[root@master ~]# oc new-app postgresql-persistent.json \
  -e DATABASE_SERVICE_NAME=postgresql \
  -e POSTGRESQL_DATABASE=testdb \
  -e POSTGRESQL_PASSWORD=testdb \
  -e POSTGRESQL_USER=testdb \
  -e VOLUME_CAPACITY=10Gi \
  -e MEMORY_LIMIT=512M \
  -e POSTGRESQL_VERSION=10 \
  -e NAMESPACE=default \
  centos/postgresql-10-centos7
```

The disk is being created in oVirt for us by the persistent storage claim:

```console
[root@master0 ~]# oc get pvc/postgresql
NAME         STATUS    VOLUME                                     CAPACITY      ACCESS MODES   STORAGECLASS   AGE
postgresql   Bound     pvc-70a8ea75-0e03-11e9-8188-001a4a160100   10737418240   RWO            ovirt          5m
```

To demonstrate that oVirt created the disk, let's look for a disk with the same name as the VOLUME name of the claim:

```console
[root@master0 ~]# curl -k  \
  -u admin@internal \
  'https://ovirt-engine-fqdn/ovirt-engine/api/disks?search=name=pvc-70a8ea75-0e03-11e9-8188-001a4a160100'  \
  | grep status
<status>ok</status>
```

Postgresql is ready with a persistent disk for its data!

Connect to the postgresql container:
```console
[root@master0 ~]# oc rsh postgresql-10-centos7-1-89ldp
```

Query the `testdb` db:
```console
sh-4.2$ psql testdb testdb -c "\l"
                                 List of databases
   Name    |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges
-----------+----------+----------+------------+------------+-----------------------
 postgres  | postgres | UTF8     | en_US.utf8 | en_US.utf8 |
 template0 | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
           |          |          |            |            | postgres=CTc/postgres
 template1 | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
           |          |          |            |            | postgres=CTc/postgres
 testdb    | testdb   | UTF8     | en_US.utf8 | en_US.utf8 |
(4 rows)

### Deploy a demo message-logger application
I wrote a demo application in java on thorntail(former wildfly-swarm), that will persist messages to the db.
We will create the application from git sources:

```console
[root@master0 ~]# oc new-app https://github.com/
```

