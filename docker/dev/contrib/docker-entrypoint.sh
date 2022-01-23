#!/usr/bin/env bash

set -e

function main {
  >&2 echo "dropping into shell"
  exec /bin/bash
}

main "$@"
