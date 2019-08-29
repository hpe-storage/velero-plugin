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

### 2. Create secret to communicate with HPE Nimble Storage CSP Service

```markdown

https://raw.githubusercontent.com/hpe-storage/csi-driver/master/examples/kubernetes/secret.yaml

```

### 3. Deploy the CSP service for HPE Nimble Storage in velero namespace

The CSP yaml file is located at [nimble-csp.yaml](nimble-csp.yaml)

### 4. Install velero HPE blockstore plugin

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
