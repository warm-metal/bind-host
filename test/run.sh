#!/usr/bin/env bash

CASE=$1

for i in 0*.sh; do
  if [[ "${CASE}" != "" ]]; then
    if [[ "${CASE}" == "$i" ]]; then
      echo "🛠 Run $(basename $i)"
      ./$i
      break
    fi
  else
    echo "🛠 Run $(basename $i)"
    ./$i
  fi
done

set +e
