# Credentials

Both `ovirt-flexvolume-driver` and `ovirt-volume-provisioner` require credentials to access the oVirt API.

The following options are available to specify the credentials (listed in decreasing order of precedence):

(NOTE: The same applies to the `ovirt-volume-provisioner`)

**Secret Injected as volume mount**
```
apiVersion: v1
kind: Secret
metadata:
  name: ovirt
data:
  credentials: <base64>
  # original file content:
  #   username=<username>
  #   password=<password>
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: ovirt-flexvolume-driver
spec:
  template:
    spec:
      containers:
      - name: ovirt-flexvolume-driver
        env:
        - name: OVIRT_CONNECTION_CREDENTIAL_FILE
          value: /opt/ovirt-credentials/credentials
        volumeMounts:
        - mountPath: /opt/ovirt-credentials
          name: secret-volume
          readOnly: true
      volumes:
      - name: secret-volume
        secret:
          secretName: ovirt
```


**Secret Injected as environment variables**
```
apiVersion: v1
kind: Secret
metadata:
  name: ovirt
data:
  username: <base64-username>
  password: <base64-password>
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: ovirt-flexvolume-driver
spec:
  template:
    spec:
      containers:
      - name: ovirt-flexvolume-driver
        env:
        - name: OVIRT_CONNECTION_USERNAME
          valueFrom:
            secretKeyRef:
              name: ovirt
              key: username
        - name: OVIRT_CONNECTION_PASSWORD
          valueFrom:
            secretKeyRef:
              name: ovirt
              key: password
```


**As part of the configmap mounted into the container**
```
apiVersion: v1
data:
  connection: |
    url=https://ovirt-engine-fqdn/ovirt-engine/api
    username=admin@internal
    password=123
    insecure=True
    cafile=
kind: ConfigMap
metadata:
  name: ovirt
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: ovirt-flexvolume-driver
spec:
  template:
    spec:
      containers:
      - name: ovirt-flexvolume-driver
        volumeMounts:
        - name: config-volume
          # must be '/etc/ovirt' for `ovirt-volume-provisioner`
          mountPath: /opt/ovirt-flexvolume-driver

      volumes:
      - configMap:
          defaultMode: 420
          items:
          - key: connection
            # must be 'ovirt-api.conf' for `ovirt-volume-provisioner`
            path: ovirt-flexvolume-driver.conf
          name: ovirt
        name: config-volume
```
