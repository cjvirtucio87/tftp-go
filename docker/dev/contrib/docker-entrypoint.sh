#!/usr/bin/env bash

set -e

function main {
  printf "%s:x:%d:%d:Developer:%s:/usr/sbin/nologin" \
    "${USER}" \
    "$(id -u)" \
    0 \
    "${HOME}" >> /etc/passwd

  >&2 echo "dropping into shell"
  exec /bin/bash
}

main "$@"
