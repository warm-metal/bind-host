package cri

import (
	"context"
	"github.com/warm-metal/bindhost/pkg/plugin"
	"google.golang.org/grpc"
	criapis "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"time"
)

func GetVolumes(criConn string) ([]plugin.MountVolume, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, criConn, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	imgSvc := criapis.NewImageServiceClient(conn)
	resp, err := imgSvc.ImageFsInfo(context.TODO(), &criapis.ImageFsInfoRequest{})
	if err != nil {
		return nil, err
	}

	volumes := make([]plugin.MountVolume, 0, len(resp.ImageFilesystems))
	for _, fs := range resp.ImageFilesystems {
		if fs.FsId == nil {
			continue
		}

		volumes = append(volumes, plugin.MountVolume{
			Source:  fs.FsId.Mountpoint,
			Target:  fs.FsId.Mountpoint,
			FsType:  "",
			Options: []string{"rbind"},
		})
	}

	return volumes, nil
}
