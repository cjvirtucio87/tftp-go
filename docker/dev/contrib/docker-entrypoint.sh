#!/usr/bin/env bash

set -e

function main {
  printf "%s:x:%d:%d:Developer:%s:/usr/bin/bash" \
    "${USER}" \
    "$(id --user)" \
    "$(id --group)" \
    "${HOME}" >> /etc/passwd

  >&2 echo "dropping into shell"
  exec /bin/bash
}

main "$@"
