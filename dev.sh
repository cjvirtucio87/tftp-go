#!/usr/bin/env bash

set -e

### Build and run the dev container (drops into a shell).
###
### Usage:
###   <Options> ./dev.sh <Arguments>

DEV_ROOT_DIR="$(dirname "$(readlink --canonicalize "$0")")"
readonly DEV_ROOT_DIR

function main {
  # shellcheck disable=SC1090,SC1091
  . "${DEV_ROOT_DIR}/build.sh"

  docker_build_dev

  docker run \
    --rm \
    --interactive \
    --tty \
    --name tftp-go-dev \
    "${DEV_IMAGE_TAG}"
}

main "$@"
