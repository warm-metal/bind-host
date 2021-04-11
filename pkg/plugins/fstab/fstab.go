package fstab

import (
	"github.com/d-tux/go-fstab"
	"github.com/warm-metal/bindhost/pkg/plugin"
)

func GetVolumes(fstabFile string) ([]plugin.MountVolume, error) {
	mounts, err := fstab.ParseFile(fstabFile)
	if err != nil {
		return nil, err
	}

	volumes := make([]plugin.MountVolume, 0, len(mounts))
	for _, m := range mounts {
		opts := make([]string, 0, len(m.VfsType))
		for k, v := range m.MntOps {
			opts = append(opts, k + "="+v)
		}

		volumes = append(volumes, plugin.MountVolume{
			Source:  m.Spec,
			Target:  m.File,
			FsType:  m.VfsType,
			Options: opts,
		})
	}

	return volumes, nil
}
