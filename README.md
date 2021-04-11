# BindHost

`BindHost` is a tool to mount volumes specified by plugins, and then, run then given command. 
Users can also provide a local folder as the base directory of all volume sources. 
In containers, we usually mount the host rootfs as the base.

It is designed for programs works in container which already has host rootfs mounted.
It provides two plugins to define volumes to be mounted, `fstab` and `cri`.

The `fstab` plugin receives a fstab(5) file and mounts volumes it defined.
The `cri` plugin fetches CRI image filesystem mountpoint via the CRI image service and mount it to the same position.

It is now used in our project [csi-driver-image](https://github.com/warm-metal/csi-driver-image).

## Usage

```shell script
Usage of _output/bind-host:
  -cri-image string
    	CRI image service endpoint. It usually is a UNIX socket URL. The image filesystem mountpoint will be retrieved via the CRI image service, then mounted to the local filesystem. (default "unix:///run/containerd/containerd.sock")
  -fstab string
    	Path of a file in the style of fstab(5). It should be absolute. Entries in the file will be mounted tothe local filesystem.
  -rootfs string
    	Path of the mounted host rootfs. It should be absolute. (default "/host")
```

Users can also use our Dockerfile as one of your build stage and copy the binary to their images.
