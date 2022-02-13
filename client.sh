#!/usr/bin/env bash

set -e

### Start the TFTP server.
###
### Usage:
###  ./client.sh

ROOT_DIR="$(dirname "$(readlink --canonicalize "$0")")"
readonly ROOT_DIR

function cleanup {
  echo "--- start cleanup ---"
  rm -f "${TEMP_FILE:?}"
  echo "--- end cleanup ---"
}

function main {
  trap cleanup EXIT

  TEMP_FILE="$(mktemp --suffix '_client')"

  go build -o "${TEMP_FILE}" cmd/client/client.go

  chmod ug+x "${TEMP_FILE}"
  "${TEMP_FILE}" -filename foo.txt
}

main
