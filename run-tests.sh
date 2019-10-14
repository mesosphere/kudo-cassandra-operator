#!/usr/bin/env bash
# shellcheck disable=SC2039

set -euxo pipefail

readonly SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly PROJECT_DIRECTORY="$(readlink -f "${SCRIPT_DIRECTORY}")"

# shellcheck source=metadata.sh
source "${PROJECT_DIRECTORY}/metadata.sh"

readonly KUBECONFIG="${KUBECONFIG:-${HOME}/.kube/config}"

readonly CONTAINER_KUBECONFIG="/root/.kube/config"
readonly CONTAINER_PROJECT_DIRECTORY="/${PROJECT_NAME}"
readonly CONTAINER_OPERATOR_DIRECTORY="${CONTAINER_PROJECT_DIRECTORY}/operator"
readonly CONTAINER_VENDOR_DIRECTORY="${CONTAINER_PROJECT_DIRECTORY}/shared/vendor"

# DS_KUDO_VERSION is used by the shared tooling.

docker run \
       --rm \
       -e "KUBECONFIG=${CONTAINER_KUBECONFIG}" \
       -e "KUBECTL_PATH=${CONTAINER_VENDOR_DIRECTORY}/kubectl.sh"  \
       -e "DS_KUDO_VERSION=v${KUDO_VERSION}" \
       -e "OPERATOR_DIRECTORY=${CONTAINER_OPERATOR_DIRECTORY}" \
       -e "VENDOR_DIRECTORY=${CONTAINER_VENDOR_DIRECTORY}" \
       -v "${KUBECONFIG}:${CONTAINER_KUBECONFIG}:ro" \
       -v "${OPERATOR_DIRECTORY}:${CONTAINER_OPERATOR_DIRECTORY}:ro" \
       -v "${PROJECT_DIRECTORY}:${CONTAINER_PROJECT_DIRECTORY}" \
       -v "${VENDOR_DIRECTORY}:${CONTAINER_VENDOR_DIRECTORY}" \
       -w "${CONTAINER_PROJECT_DIRECTORY}" \
       "${INTEGRATION_TESTS_DOCKER_IMAGE}" \
       bash -c "${CONTAINER_PROJECT_DIRECTORY}/tests/run.sh"
