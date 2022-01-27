#!/usr/bin/env bash

set -eo pipefail

### Build and run the dev container (drops into a shell).
###
### Usage:
###   <Options> ./dev.sh <Arguments>

DEV_ROOT_DIR="$(dirname "$(readlink --canonicalize "$0")")"
readonly DEV_ROOT_DIR
readonly DEV_CONTAINER_NAME='tftp-go-dev'

function cleanup {
  >&2 echo "starting cleanup"
  docker stop "${DEV_CONTAINER_NAME}" || true
  >&2 echo "done cleanup"
}

function main {
  >&2 echo "preliminary cleanup"
  cleanup
  trap cleanup EXIT

  # shellcheck disable=SC1090,SC1091
  . "${DEV_ROOT_DIR}/build.sh"

  docker_build_dev

  docker run \
    --rm \
    --interactive \
    --tty \
    --mount "type=bind,src=${DEV_ROOT_DIR},dst=/home/dev" \
    --user "$(id --user):0" \
    --name "${DEV_CONTAINER_NAME}" \
    "${DEV_IMAGE_TAG}"
}

main "$@"
