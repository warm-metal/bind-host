# BindHost

`BindHost` is a tool to mount host dirs/files into containers, and then, run the given command,
and unmount them before exiting.

It usually works in containers as entrypoints. Users should mount the host rootfs into the container,
say in directory `/host`, `BindHost` will mount volumes plugins specified from `/host` to the local filesystem. 

If a command is given, it will share stdin, stderr and stdout with `BindHost`.
`BindHost` also returns the exit code of the command. 

It is now used in our project [csi-driver-image](https://github.com/warm-metal/csi-driver-image).

## Plugins
`BindHost` provides two plugins to define volumes to be mounted, `fstab` and `cri`.

* The `fstab` plugin receives a fstab(5) file and mounts volumes it defined.
* The `cri` plugin fetches CRI image filesystem mountpoint via the CRI image service and mount it to the same position.

## Usage

#### CRI Image
Mount the CRI image mountpoint from `/host` to the local filesystem and then run the CSI binary `/csi-image-plugin`.

Say containerd, the image mountpoint usually is `/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs`.
`bind-host` will mount **/host**`/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs` to `/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs`.
 
```shell script
bind-host -rootfs=/host -cri-image=unix:///run/containerd/containerd.sock -- /csi-image-plugin
```

#### Mount specified volumes and wait for a signal
Mount directory `/host/etc/systemd` to `/foo/bar` and wait for a signal to unmount it, then exit.

The flag `-v=1` enables verbose logs, or nothing print unless errors arise.

```shell script
echo '/etc/systemd /foo/bar none defaults,rbind' > ./fstab
bind-host -rootfs=/host -fstab=./fstab -v 1 -wait
```

## Integration

### Download the binary
(We haven't released yet!)

Users can download prebuilt binaries from the release page.

### Embed the Dockerfile
Users can also use our `Dockerfile.embeded` as one of your build stage and copy the binary to their images.

### Copy the binary from our image
Users can use our published image `docker.io/warmmetal/bind-host:latest` as the based image of one stage.
