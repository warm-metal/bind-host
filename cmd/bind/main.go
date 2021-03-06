package main

import (
	"flag"
	"fmt"
	"github.com/warm-metal/bindhost/pkg/plugin"
	"github.com/warm-metal/bindhost/pkg/plugins/cri"
	"github.com/warm-metal/bindhost/pkg/plugins/fstab"
	"k8s.io/klog/v2"
	"k8s.io/utils/mount"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	rootfs    = ""
	criConn   = ""
	fstabPath = ""
	verbosity = ""
	waitSignal = false
)

func init() {
	flag.StringVar(&rootfs, "rootfs", rootfs, "Path of the mounted host rootfs. It should be absolute. [$HOST_ROOTFS]")
	flag.StringVar(&criConn, "cri-image", criConn,
		"CRI image service endpoint. It usually is a UNIX socket URL. The image filesystem mountpoint will be "+
			"retrieved via the CRI image service, then mounted to the local filesystem. [$CRI_ADDR]",
	)
	flag.StringVar(&fstabPath, "fstab", "",
		"Path of a file in the manner of fstab(5). It should be absolute. Entries in the file will be mounted to "+
			"the local filesystem. [$FSTAB]",
	)
	flag.StringVar(&verbosity, "v","", "Number for the log level verbosity. Set to 1 to show debug logs.")
	flag.BoolVar(&waitSignal, "wait", false,
		"Wait for signal SIGTERM, SIGQUIT or SIGINT to exit if no command given.",
	)
}

func main() {
	flag.CommandLine.Usage = printUsage
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}

		fmt.Printf("\ncommand line got: %s\n", strings.Join(os.Args, " "))
		os.Exit(2)
	}

	// Parse flags frm envs.
	envFlags := map[string]*string{
		"HOST_ROOTFS": &rootfs,
		"CRI_ADDR": &criConn,
		"FSTAB": &fstabPath,
	}

	for k, f := range envFlags {
		if len(*f) > 0 {
			continue
		}

		*f = os.Getenv(k)
	}

	klogFlags := flag.NewFlagSet("klog", flag.PanicOnError)
	klog.InitFlags(klogFlags)
	klogFlags.Set("logtostderr", "true")
	klogFlags.Set("v", verbosity)
	klogFlags.Parse(nil)

	defer klog.Flush()

	klog.V(1).Infof("debug logs enabled!")

	if !filepath.IsAbs(rootfs) {
		klog.Fatalf("-rootfs must be absolute but %q", rootfs)
		return
	}

	if _, err := os.Lstat(rootfs); err != nil {
		klog.Fatal(err.Error())
		return
	}

	mounter := mount.New("")
	var volumes []plugin.MountVolume
	if len(criConn) > 0 {
		vs, err := cri.GetVolumes(criConn)
		if err != nil {
			klog.Fatalf("cri plugin failed: %s", err)
		}

		volumes = append(volumes, vs...)
	}

	if len(fstabPath) > 0 {
		vs, err := fstab.GetVolumes(fstabPath)
		if err != nil {
			klog.Fatalf("fstab plugin failed: %s", err)
		}

		volumes = append(volumes, vs...)
	}

	signCh := make(chan os.Signal, 3)
	defer close(signCh)
	signal.Notify(signCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	var cmd *exec.Cmd
	if len(flag.Args()) > 0 {
		cmd = exec.Command(flag.Arg(0), flag.Args()[1:]...)
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	var signExited chan struct{}
	if waitSignal || cmd != nil {
		signExited = make(chan struct{})
		go func() {
			defer close(signExited)
			sig := <-signCh
			signal.Stop(signCh)
			klog.V(1).Infof("got signal %s", sig)

			if cmd != nil {
				if cmd.Process != nil {
					cmd.Process.Signal(sig)
				}
				return
			}

			return
		}()
	}

	defer unmount(mounter, volumes)

	for _, v := range volumes {
		src := filepath.Join(rootfs, v.Source)
		srcStat, err := os.Lstat(src)
		if err != nil {
			panic(err.Error())
		}

		_, err = os.Lstat(v.Target)
		if err != nil {
			if !os.IsNotExist(err) {
				panic(err.Error())
			}

			if srcStat.IsDir() {
				if err := os.MkdirAll(v.Target, srcStat.Mode()); err != nil {
					klog.Fatalf("unable to mkdir %q(%#o): %s", v.Target, srcStat.Mode(), err)
				}
			} else {
				if err := os.MkdirAll(filepath.Dir(v.Target), srcStat.Mode()); err != nil {
					klog.Fatalf("unable to mkdir %q(%#o): %s", filepath.Dir(v.Target), srcStat.Mode(), err)
				}

				f, err := os.Create(v.Target)
				if err != nil {
					klog.Fatalf("unable to touch %q: %s", v.Target, err)
				}
				f.Close()
			}
		}

		if err := mounter.Mount(src, v.Target, v.FsType, v.Options); err != nil {
			klog.Fatalf("unable to mount volume %q: %s", v, err)
		}

		klog.V(1).Infof("mount %q to %q\n", src, v.Target)
	}

	if cmd != nil {
		klog.V(1).Infof("exec %#v", flag.Args())
		err := cmd.Run()
		klog.V(1).Infof("exec done with err %#v", err)
		if err != nil {
			klog.Error(err.Error())
			if st, exit := err.(*exec.ExitError); exit {
				unmount(mounter, volumes)
				os.Exit(st.ProcessState.ExitCode())
			}
		}

		return
	}

	if signExited != nil {
		<-signExited
	}
}

func unmount(mounter mount.Interface, volumes []plugin.MountVolume) {
	for i := len(volumes) - 1; i >= 0; i-- {
		v := volumes[i]
		klog.V(1).Infof("unmount %q\n", v.Target)
		if err := mounter.Unmount(v.Target); err != nil {
			klog.Errorf("unable to unmount %q: %s\n", v.Target, err)
		}
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr,`
bind-host [Flags] -- [command args]

bind-host mounts directories or files from the path given via -rootfs to the local filesystem. If command and args are given, it will be executed after mounting all volumes.

Flags:`)
	flag.PrintDefaults()
}
