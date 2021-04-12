#!/usr/bin/env bash

set -e
set -x

echo '/etc/systemd /foo/bar none defaults,rbind' > ./fstab
MountpointID=$(./bind-host -rootfs=${HOST_ROOTFS} -fstab=./fstab -v 1 -- stat -c%i /foo/bar)
SrcID=$(stat -c%i /host/etc/systemd)
TargetID=$(stat -c%i /foo/bar)

if [ $MountpointID -ne $SrcID ]; then
  echo "mount failed"
  exit 1
fi

if [ $MountpointID -eq $TargetID ]; then
  echo "unmount failed"
  exit 1
fi

set +x
set +e