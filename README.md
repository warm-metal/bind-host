#BindHost

`BindHost` is a tool to find out the CRI image filesystem and mount them from host rootfs.

It is designed to used in containers which already mount the host rootfs and need to work with images stored in various container runtime,
such as the [csi-driver-image](https://github.com/warm-metal/csi-driver-image).

## Usage

The `bind-host` binary has 2 flags, `-rootfs` and `-cri-image`, for respectively the mountpoint of host rootfs and
the host CRI endpoint which usually is a unix socket URL.

You can also use our Dockerfile as one of your build stage and copy the binary to your image.