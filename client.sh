#!/usr/bin/env bash

set -e

### Communicate with a TFTP server.
###
### Usage:
###  ./client.sh

ROOT_DIR="$(dirname "$(readlink --canonicalize "$0")")"
readonly ROOT_DIR

function log {
  >&2 printf '[%s] %s\n' "$(date --iso=s)" "$1"
}

function cleanup {
  log "--- start cleanup ---"
  rm -f "${TEMP_FILE:?}"
  log "--- end cleanup ---"
}

function main {
  trap cleanup EXIT

  TEMP_FILE="$(mktemp --suffix '_client')"

  go build -o "${TEMP_FILE}" cmd/client/client.go

  chmod ug+x "${TEMP_FILE}"
  "${TEMP_FILE}" -filename foo.txt
}

main
