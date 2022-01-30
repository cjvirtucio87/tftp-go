#!/usr/bin/env bash

set -eo pipefail

### Build and run the dev container (drops into a shell).
###
### Usage:
###   ./dev.sh <Arguments>
###
### Arguments:
###   local_workspace_folder: path to the local workspace
###     that should be mounted into the container

DEV_ROOT_DIR="$(dirname "$(readlink --canonicalize "$0")")"
readonly DEV_ROOT_DIR
readonly DEV_CONTAINER_NAME='vim-go-devc'

function cleanup {
  >&2 echo "starting cleanup"
  docker stop "${DEV_CONTAINER_NAME}" || true
  >&2 echo "done cleanup"
}

function main {
  local local_workspace_folder="${DEV_ROOT_DIR}"

  >&2 echo "preliminary cleanup"
  cleanup
  trap cleanup EXIT

  # shellcheck disable=SC1090,SC1091
  . "${DEV_ROOT_DIR}/build.sh"

  docker_build_vim_go_devc
  local workspace_name
  workspace_name="$(basename "${local_workspace_folder}")"
  local container_workspace_folder="/workspaces/${workspace_name}"
  docker run \
    --rm \
    --interactive \
    --tty \
    --env "AD_HOC=1" \
    --mount "type=bind,src=${local_workspace_folder},dst=${container_workspace_folder}" \
    --workdir "${container_workspace_folder}" \
    --user "$(id --user):$(id --group)" \
    --name "${DEV_CONTAINER_NAME}" \
    "${VIM_GO_DEVC_IMAGE_TAG}"
}

main "$@"
