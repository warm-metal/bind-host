package main

import (
	"flag"
	"fmt"
	"github.com/warm-metal/bindhost/pkg/plugin"
	"github.com/warm-metal/bindhost/pkg/plugins/cri"
	"github.com/warm-metal/bindhost/pkg/plugins/fstab"
	"k8s.io/utils/mount"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	rootfs    = "/host"
	criConn   = "unix:///run/containerd/containerd.sock"
	fstabPath = ""
)

func init() {
	flag.StringVar(&rootfs, "rootfs", rootfs, "Path of the mounted host rootfs. It should be absolute.")
	flag.StringVar(&criConn, "cri-image", criConn,
		"CRI image service endpoint. It usually is a UNIX socket URL. The image filesystem mountpoint will be "+
			"retrieved via the CRI image service, then mounted to the local filesystem.",
	)
	flag.StringVar(&fstabPath, "fstab", "",
		"Path of a file in the style of fstab(5). It should be absolute. Entries in the file will be mounted to "+
			"the local filesystem.",
	)
}

func main() {
	flag.Parse()
	if !filepath.IsAbs(rootfs) {
		printUsage()
		return
	}

	mounter := mount.New("")
	var volumes []plugin.MountVolume
	if len(criConn) > 0 {
		var err error
		vs, err := cri.GetVolumes(criConn)
		if err != nil {
			panic(fmt.Sprintf("cri plugin failed: %s", err))
		}

		volumes = append(volumes, vs...)
	}

	if len(fstabPath) > 0 {
		vs, err := fstab.GetVolumes(fstabPath)
		if err != nil {
			panic(fmt.Sprintf("fstab plugin failed: %s", err))
		}

		volumes = append(volumes, vs...)
	}

	defer func() {
		for i := len(volumes) - 1; i >= 0; i-- {
			v := volumes[i]
			if err := mounter.Unmount(v.Target); err != nil {
				fmt.Fprintln(os.Stderr, "unable to unmount %q: %s", v.Target, err)
			}
		}
	}()

	for _, v := range volumes {
		if err := mounter.Mount(filepath.Join(rootfs, v.Source), v.Target, v.FsType, v.Options); err != nil {
			panic(fmt.Sprintf("unable to mount volume %q: %s", v, err))
		}
	}

	if len(flag.Args()) > 0 {
		cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err.Error())
			if st, exit := err.(*exec.ExitError); exit {
				os.Exit(st.ProcessState.ExitCode())
			}
		}
	}
}

func printUsage() {
	fmt.Print(`bind-host [Flags] -- [command args]

bind-host mounts directories or files from the path given via -rootfs to the local rootfs. If command and args are given,
the command will be executed after all volumes mounted. Or, bind-host will exit.

`)
	flag.CommandLine.Usage()
}
