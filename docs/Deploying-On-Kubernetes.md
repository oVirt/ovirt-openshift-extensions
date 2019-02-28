Volume provisioner and flexvolume driver can easily be deployed on kubernetes cluster.
Following document explains quick steps to deploy ovirt-openshift-extensions on kubernetes enviroment.

## Deployment

1. Checkout sample deployment yaml's from https://gist.github.com/rgolangh/b8eb941ece1d7b494b232a80aabfb2d6

2. Update configmap.yaml to provide your ovirt connections details. Mainly you need to modify url, password fields in the configmap.yaml
    Create config map by:
    ```console
    kubectl create -f configmap.yaml
    ```

3. Create cluster role, role bindings & service account required for volume provisioner and flexvoume driver.
   ```console
    kubectl create -f rbac-all-in-one.yaml
    ```

4. Create volume provisoner deployment.
   ```console
   kubectl create -f provisioner.yaml
   ```
   Confirm that ovirt-volume-provisioner pod is scheduled and up by running `kubectl get pods`.

5. Deploy flexvolume driver on all compute nodes by running following on the master node.
   ```console
   kubectl create -f flexvolume.yaml
   ````
   You can confirm that ovirt-flexvolume-driver pod is scheduled and up on all `kubernetes get pods`.
   Please check the logs for provisioner &  flexvolume-driver pods to confirm it can access ovirt url using 'kubectl get logs YOUR-POD-NAME'.
   If  kubenetes compute node has problem accessing ovirt engine url, you might need to tweak your kubernetes cluster
   network configs or provide custom dns settings in flexvoume.yaml & provisoner.yaml and recreate them.
   ```console
       dnsConfig:
         nameservers:
           - YOUR-DNS-SERVER-IP
    ```


Now your kubernetes cluster is ready to provision volumes from ovirt.

## Test

Test your cluster by checking out https://github.com/oVirt/ovirt-openshift-extensions/blob/master/deployment/example/test-flex-pod-all-in-one.yaml
Update `ovirtStorageDomain` field in `test-flex-pod-all-in-one.yaml` file to refer to your ovirt storage domain name on which
volume needs to be provisioned.

Then spin off testpodwithflex by running
```console
kubectl create -f test-flex-pod-all-in-one.yaml
```
Verify that volume is provisioned on your strorage domain and it's attached compute node where testpodwithflex is scheduled.

Enjoy your setup. Your feedback and contributions are most welcome.
