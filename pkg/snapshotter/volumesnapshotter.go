// Copyright 2019 Hewlett Packard Enterprise Development LP

package snapshotter

import (
	"fmt"
	"github.com/hpe-storage/common-host-libs/storageprovider"
	"github.com/hpe-storage/common-host-libs/storageprovider/csp"
	"github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type volumesnapshotter struct {
	Snapshotter
	Log             logrus.FieldLogger
	storageProvider *csp.ContainerStorageProvider
}

func (v *volumesnapshotter) Init(config map[string]string) error {
	v.Log.Infof(">>> Init hpe volumesnapshotter with config %+v", config)
	defer v.Log.Infof("<<<< Init")
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("error getting config cluster - %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return fmt.Errorf("error getting client - %s", err.Error())
	}
	secret, err := kubeClient.CoreV1().Secrets(config[secretNamespaceKey]).Get(config[secretNameKey], meta_v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting secret - %s", err.Error())

	}
	secrets := map[string]string{}
	for key, value := range secret.Data {
		secrets[key] = string(value)
	}
	credentials, err := storageprovider.CreateCredentials(secrets)
	if err != nil {
		return fmt.Errorf("error connecting to storage provider - %s", err.Error())
	}
	storageProvider, err := csp.NewContainerStorageProvider(credentials)
	if err != nil {
		return fmt.Errorf("error building storage provider - %s", err.Error())
	}
	v.storageProvider = storageProvider
	return nil
}

// CreateVolumeFromSnapshot creates a new volume initialized from the provided snapshot,
// and with the specified type and IOPS (if using provisioned IOPS).
func (v *volumesnapshotter) CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ string, iops *int64) (string, error) {
	v.Log.Infof(">>>>> CreateVolumeFromSnapshot called with snapshot ID %s, volume type %s, volumeAZ %s, and iops %d",
		snapshotID, volumeType, volumeAZ, iops)
	defer v.Log.Infof("<<<<< CreateVolumeFromSnapshot")

	snapshot, err := v.storageProvider.GetSnapshot(snapshotID)
	if err != nil {
		return "", err
	}
	if snapshot.Name == "" {
		return "", fmt.Errorf("unable to retrieve snapshot with id %s", snapshotID)
	}
	volume, err := v.storageProvider.CloneVolume("clone-of-"+snapshot.Name, "Clone request from velero", "", snapshotID, 0, make(map[string]interface{}))
	if err != nil {
		v.Log.Infof("Failed to clone volume - %s", err.Error())
		return "", err
	}
	return volume.ID, err
}

// GetVolumeInfo returns the type and IOPS (if using provisioned IOPS) for
// the specified volume in the given availability zone.
func (v *volumesnapshotter) GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error) {
	v.Log.Infof(">>>>> GetVolumeInfo called with id %s and name %s", volumeID, volumeAZ)
	defer v.Log.Infof("<<<<< GetVolumeInfo")

	// we do not store the type or iops
	return "", nil, nil
}

// CreateSnapshot creates a snapshot of the specified volume, and applies any provided
// set of tags to the snapshot.
func (v *volumesnapshotter) CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error) {
	v.Log.Infof(">>>>> CreateSnapshot for volume ID %s with volumeAZ %s and tags %+v", volumeID, volumeAZ, tags)
	defer v.Log.Infof("<<<<< CreateSnapshot")

	snapName := "velero"
	if val, ok := tags[veleroBackupKey]; ok {
		snapName += "-" + val
	}

	// create a snapshot
	snapshot, err := v.storageProvider.CreateSnapshot(snapName, "snapshot from velero", volumeID, nil)
	if err != nil {
		v.Log.Infof("Failed to create snapshot %s - %s", snapName, err.Error())
		return "", err
	}
	return snapshot.ID, err
}

// DeleteSnapshot deletes the specified volume snapshot.
func (v *volumesnapshotter) DeleteSnapshot(snapshotID string) error {
	v.Log.Infof(">>>>> DeleteSnapshot with id %s", snapshotID)
	defer v.Log.Infof("<<<<< DeleteSnapshot")
	return v.storageProvider.DeleteSnapshot(snapshotID)
}
