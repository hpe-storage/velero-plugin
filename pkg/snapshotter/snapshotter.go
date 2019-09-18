// Copyright 2019 Hewlett Packard Enterprise Development LP

package snapshotter

import (
	"fmt"
	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	secretNameKey      = "secret-name"
	secretNamespaceKey = "secret-namespace"
	veleroBackupKey    = "velero.io/backup"
)

// Snapshotter is a plugin for containing state for the blockstore
type Snapshotter struct {
	Log    logrus.FieldLogger
	plugin velero.VolumeSnapshotter
}

// Init prepares the Snapshotter for usage using the provided map of
// configuration key-value pairs. It returns an error if the Snapshotter
// cannot be initialized from the provided config.
func (s *Snapshotter) Init(config map[string]string) error {
	s.Log.Infof(">>>>> Init snapshotter with config %v", config)
	defer s.Log.Infof("<<<<< Init snapshotter")
	s.plugin = &volumesnapshotter{Log: s.Log}
	return s.plugin.Init(config)
}

// GetVolumeID Get the volume ID from the spec
func (s *Snapshotter) GetVolumeID(unstructuredPV runtime.Unstructured) (string, error) {
	s.Log.Infof(">>>>> GetVolumeID called with unstructuredPV %v", unstructuredPV)
	defer s.Log.Infof("<<<<< GetVolumeID")
	pv := new(v1.PersistentVolume)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredPV.UnstructuredContent(), pv); err != nil {
		return "", errors.WithStack(err)
	}
	if pv.Spec.CSI == nil {
		return "", fmt.Errorf("unable to retrieve CSI Spec from pv %+v", pv)
	}
	if pv.Spec.CSI.VolumeHandle == "" {
		return "", fmt.Errorf("unable to retrieve Volume handle from pv %+v", pv)
	}
	return pv.Spec.CSI.VolumeHandle, nil
}

// SetVolumeID Set the volume ID in the spec
func (s *Snapshotter) SetVolumeID(unstructuredPV runtime.Unstructured, volumeID string) (runtime.Unstructured, error) {
	s.Log.Infof(">>>>> SetVolumeID called with unstructuredPV %v and volumeID %s", unstructuredPV, volumeID)
	defer s.Log.Infof("<<<<< SetVolumeID")
	pv := new(v1.PersistentVolume)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredPV.UnstructuredContent(), pv); err != nil {
		return nil, errors.WithStack(err)
	}

	if pv.Spec.CSI == nil {
		return nil, fmt.Errorf("spec.CSI not found from pv %+v", pv)
	}

	pv.Spec.CSI.VolumeHandle = volumeID

	res, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pv)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &unstructured.Unstructured{Object: res}, nil
}

// CreateVolumeFromSnapshot Create a volume form given snapshot
func (s *Snapshotter) CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ string, iops *int64) (string, error) {
	return s.plugin.CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ, iops)
}

// GetVolumeInfo Get information about the volume
func (s *Snapshotter) GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error) {
	return s.plugin.GetVolumeInfo(volumeID, volumeAZ)
}

// CreateSnapshot Create a snapshot
func (s *Snapshotter) CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error) {
	return s.plugin.CreateSnapshot(volumeID, volumeAZ, tags)
}

// DeleteSnapshot Delete a snapshot
func (s *Snapshotter) DeleteSnapshot(snapshotID string) error {
	return s.plugin.DeleteSnapshot(snapshotID)
}
