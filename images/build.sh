#!/usr/bin/env bash
# shellcheck disable=SC2039

set -euxo pipefail

readonly SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly PROJECT_DIRECTORY="$(readlink -f "${SCRIPT_DIRECTORY}/..")"

# shellcheck source=../metadata.sh
source "${PROJECT_DIRECTORY}/metadata.sh"

docker image build \
       -t "${CASSANDRA_DOCKER_IMAGE}" \
       -f "${PROJECT_DIRECTORY}/images/Dockerfile" \
       "${PROJECT_DIRECTORY}/images"

if [[ "${1:-}" == "push" ]]; then
  docker push "${CASSANDRA_DOCKER_IMAGE}"
fi
