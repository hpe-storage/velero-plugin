# velero-plugin

HPE plugin for Velero.
To take snapshots of HPE volumes through velero you need to install and configure the HPE Snapshotter plugin.

## Installation

Refer to <https://velero.io/docs/v1.1.0/get-started/> for installing velero.

### 1. Create VolumeSnapshotLocation CRD for HPE Snapshotter

```markdown

apiVersion: velero.io/v1
kind: VolumeSnapshotLocation
metadata:
  name: hpe-csp
  namespace: velero
spec:
  provider: hpe.com/snapshotter
  config: {
       "secret-name": "nimble-secret",
       "secret-namespace": "kube-system"
  }
```

### 2. Deploy a Container Service Provider service

In order for the snapshotter to perform snapshots, it needs to communicate with a CSP.
Currently only HPE Nimble Storage provides a CSP and the rest of the installation assumes Nimble.

Create a CSP secret that maps to a Nimble array management IP address and "Power User" (or Administrator):

```markdown

apiVersion: v1
kind: Secret
metadata:
  name: nimble-secret
  namespace: kube-system
stringData:
  serviceName: nimble-csp-svc
  servicePort: "8080"
  backend: 192.168.1.1
  username: admin
data:
  # echo -n admin | base64
  password: YWRtaW4=

```

Deploy the CSP (it will be deployed in the `velero` namespace on the cluster)

```kubectl create -f https://raw.githubusercontent.com/hpe-storage/velero-plugin/master/nimble-csp.yaml```

### 3. Install Velero HPE blockstore plugin

```markdown

velero plugin add hpestorage/velero-hpe-blockstore:beta --image-pull-policy Always

```

## Backup

Everytime a velero backup is taken and include PVCs, it will also take HPE Nimble Storage snapshots of your volumes.

```markdown

velero backup create default-ns-hpe-backup --include-namespaces=default --snapshot-volumes --volume-snapshot-locations hpe-csp

Backup request "default-ns-hpe-backup" submitted successfully.
Run `velero backup describe default-ns-hpe-backup` for more details.

```

## Listing Backup

```markdown

velero backup get
NAME                      STATUS                       CREATED                         EXPIRES   STORAGE LOCATION   SELECTOR
default-ns-hpe-backup     Completed                    2019-08-16 11:49:33 -0700 PDT   17d       default            <none>
default-ns-hpe-backup2    Completed                    2019-08-19 10:09:41 -0700 PDT   20d       default            <none>
default-ns-hpe-backup3    Completed                    2019-08-16 12:13:50 -0700 PDT   17d       default            <none>

```

## Restore

Restoring from velero backup, a HPE Nimble Storage clone volume will be created from the snapshot and bound to the restored PVC. To restore from the backup created above you can run the following command:

```markdown

velero restore create --from-backup default-ns-hpe-backup

Restore request "default-ns-hpe-backup-20190828193501" submitted successfully.

```
