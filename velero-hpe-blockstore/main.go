// Copyright 2019 Hewlett Packard Enterprise Development LP

package main

import (
	veleroplugin "github.com/heptio/velero/pkg/plugin/framework"
	"github.com/hpe-storage/velero-plugin/pkg/snapshotter"
	"github.com/sirupsen/logrus"
)

func main() {

	veleroplugin.NewServer().
		RegisterVolumeSnapshotter("hpe.com/snapshotter", newSnapshotter).
		Serve()
}

func newSnapshotter(logger logrus.FieldLogger) (interface{}, error) {
	return &snapshotter.Snapshotter{Log: logger}, nil
}
