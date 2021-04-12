#!/usr/bin/env bash

set -e
set -x

echo '/etc/systemd /foo/bar none defaults,rbind' > ./fstab
./bind-host -rootfs=${HOST_ROOTFS} -fstab=./fstab -v 1 -wait &
PID=$!
sleep .5
kill $PID

set +e
wait $PID

set -e
SrcID=$(stat -c%i /host/etc/systemd)
TargetID=$(stat -c%i /foo/bar)

if [ $SrcID -eq $TargetID ]; then
  echo "unmount failed"
  exit 1
fi

set +x
set +e