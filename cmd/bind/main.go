package main

import (
	"flag"
	"fmt"
	"github.com/warm-metal/bindhost/pkg/plugin"
	"github.com/warm-metal/bindhost/pkg/plugins/cri"
	"k8s.io/utils/mount"
	"path/filepath"
)

var (
	rootfs = "/host"
	criConn = "unix:///run/containerd/containerd.sock"
)

func init() {
	flag.StringVar(&rootfs, "rootfs", rootfs, "Path of the mounted host rootfs")
	flag.StringVar(&criConn, "cri-image", criConn, "CRI image service endpoint")
}

func main() {
	flag.Parse()
	if !filepath.IsAbs(rootfs) {
		panic(fmt.Sprintf("rootfs %q must be absolute", rootfs))
	}

	mounter := mount.New("")
	var volumes []plugin.MountVolume
	if len(criConn) > 0 {
		var err error
		volumes, err = cri.GetVolumes(criConn)
		if err != nil {
			panic(fmt.Sprintf("cri plugin failed: %s", err))
		}
	}

	for _, v := range volumes {
		if err := mounter.Mount(filepath.Join(rootfs, v.Source), v.Target, v.FsType, v.Options); err != nil {
			panic(fmt.Sprintf("unable to mount volume %q: %s", v, err))
		}
	}
}