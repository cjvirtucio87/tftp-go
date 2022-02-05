#!/usr/bin/env bash

set -e

### Build images for this project.
###
### Usage:
###   <Options> ./build.sh <Arguments>

if (return 0 2>/dev/null); then
  if [[ -v _BUILD_SOURCED_ ]]; then
    >&2 echo "build.sh already sourced"
    return
  fi

  _BUILD_SOURCED_=1
fi

BUILD_ROOT_DIR="$(dirname "$(readlink --canonicalize "$0")")"
readonly BUILD_ROOT_DIR
readonly ROCKY_IMAGE_TAG="cjvirtucio87/docker-base-rockylinux:latest"

function docker_build {
  local filepath=$1
  local tag=$2

  docker build \
    --tag "${tag}" \
    --file "${filepath}" \
    "${BUILD_ROOT_DIR}"
}

function docker_build_rocky {
  docker_build "${BUILD_ROOT_DIR}/docker/rockylinux/Dockerfile" "${ROCKY_IMAGE_TAG}"
}

function main {
  docker_build_rocky
}

(return 0 2>/dev/null) || main "$@"
