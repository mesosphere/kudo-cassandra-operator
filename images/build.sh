#!/usr/bin/env bash
# shellcheck disable=SC2039

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}/..")"

# shellcheck source=../metadata.sh
source "${project_directory}/metadata.sh"

readonly cassandra_docker_image="${CASSANDRA_DOCKER_IMAGE:-}"

if [[ -z ${cassandra_docker_image} ]]; then
  echo "Missing CASSANDRA_DOCKER_IMAGE" >&2
  exit 1
fi

docker image build \
       -t "${cassandra_docker_image}" \
       -f "${project_directory}/images/Dockerfile" \
       "${project_directory}/images"

if [[ "${1:-}" == "push" ]]; then
  docker push "${cassandra_docker_image}"
fi
