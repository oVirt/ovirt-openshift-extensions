# Flexvolume Setup

## Flexvolume plugin directory

Depending on the configuration of your Kubernetes cluster some adjustments may be necessary.

Some Kubernetes distributions do not use the default flexvolume directory (`/usr/libexec/kubernetes/kubelet-plugins/volume/exec/`). In that case the `ovirt-flexvolume-driver` DaemonSet must use the correct `hostPath`:

```yaml
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: ovirt-flexvolume-driver
spec:
  template:
    spec:
      volumes:
      - hostPath:
          path: <PATH_TO_FLEXVOLUME_PLUGIN_DIR>
        name: plugindir
```


Furthermore, ensure that`kubelet` and `kube-controller-manager` are set up to use the correct plugin directory:

* `kubelet` -> command line flag `--volume-plugin-dir` ([reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/))
* `kube-controller-manager` -> command line flag `--flex-volume-plugin-dir string` ([reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/))

If kubelet or kube-controller-manager run in a container then the flexvolume directory must be mounted into the container. In this case it is not necessary to adjust the command line flags as we mount the flexvolume to the default directory in the container. For example if your flexvolume path is `/var/lib/kubelet/volume-plugins` then the kube-controller-manager pod should look like the following:

```
apiVersion: v1
kind: Pod
metadata:
  name: kube-controller-manager
  namespace: kube-system
spec:
  containers:
  - name: kube-controller-manager
    volumeMounts:
    - mountPath: /usr/libexec/kubernetes/kubelet-plugins/volume/exec
      name: flexvolume-dir
  volumes:
  - hostPath:
      path: /var/lib/kubelet/volume-plugins
    name: flexvolume-dir
```

## `kubeconfig` for ovirt-flexvolume-driver

Additionally it may be necessary to set the `KUBECONFIG` environment variable (pointing to the `kubeconfig` file that contains information on how to contact the kubernetes API server) where `kube-controller-manager` is executed. For example, if `kube-controller-manager` is executed in a container and the `kubeconfig` file is located under `/etc/kubernetes/controller-manager.conf` in the container:

```
apiVersion: v1
kind: Pod
metadata:
  name: kube-controller-manager
  namespace: kube-system
spec:
  containers:
  - name: kube-controller-manager
    env:
    - name: KUBECONFIG
      value: "/etc/kubernetes/controller-manager.conf"
```

Note: Above requirement is specific to the ovirt-flexvolume-driver implementation and is mainly
required to identify the node systemUUID, which is the underneath VM ID, to attach the disk.

