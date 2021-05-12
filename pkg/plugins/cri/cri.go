package cri

import (
	"context"
	"github.com/warm-metal/bindhost/pkg/plugin"
	"google.golang.org/grpc"
	criapis "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"
	"net/url"
	"strings"
	"time"
)

func GetVolumes(criConn string) ([]plugin.MountVolume, error) {
	addr, err := url.Parse(criConn)
	if err != nil {
		return nil, err
	}

	isCRIO := addr.Scheme == "cri-o"
	addr.Scheme = "unix"

	klog.Info("use cri socket ", addr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr.String(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	imgSvc := criapis.NewImageServiceClient(conn)
	resp, err := imgSvc.ImageFsInfo(context.TODO(), &criapis.ImageFsInfoRequest{})
	if err != nil {
		return nil, err
	}

	var vols []string
	if isCRIO {
		if vols, err = fetchCriOVolumes(addr.Path); err != nil {
			return nil, err
		}

		klog.Info("volumes of container runtime cri-o: %#v", vols)
	}

	for _, fs := range resp.ImageFilesystems {
		if fs.FsId == nil {
			continue
		}

		for i, vol := range vols {
			if strings.HasPrefix(fs.FsId.Mountpoint, vol) {
				break
			}

			if strings.HasPrefix(vol, fs.FsId.Mountpoint) {
				vols[i] = fs.FsId.Mountpoint
				break
			}

			vols = append(vols, fs.FsId.Mountpoint)
		}
	}

	volumes := make([]plugin.MountVolume, 0, len(vols))
	for _, vol := range vols {
		volumes = append(volumes, plugin.MountVolume{
			Source:  vol,
			Target:  vol,
			FsType:  "",
			Options: []string{"rbind"},
		})
	}

	return volumes, nil
}
