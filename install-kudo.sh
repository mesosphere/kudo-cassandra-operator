#!/usr/bin/env bash
# shellcheck disable=SC2039

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}")"

# shellcheck source=metadata.sh
source "${project_directory}/metadata.sh"

readonly kubeconfig="${KUBECONFIG:-${HOME}/.kube/config}"

readonly container_kubeconfig="/root/.kube/config"
readonly container_project_directory="/${PROJECT_NAME}"
readonly container_tooling_directory="${container_project_directory}/shared"
readonly container_vendor_directory="${container_project_directory}/shared/vendor"

# Note: DS_KUDO_VERSION is used by the shared data-services-kudo tooling.
# DS_KUDO_VERSION *may* be set by TeamCity Jobs if a fixed KUDO version is preferred for the test execution
# If not DS_KUDO_VERSION is set, we use and install the required KUDO version from the operator
export DS_KUDO_VERSION="${DS_KUDO_VERSION:-${KUDO_VERSION}}"

# Strip 'v' from "v0.10.1" if it's there.
if [[ ${DS_KUDO_VERSION:0:1} == "v" ]] ; then DS_KUDO_VERSION="${DS_KUDO_VERSION:1}"; fi

docker run \
       --rm \
       -e "KUBECONFIG=${container_kubeconfig}" \
       -e "KUBECTL_PATH=${container_vendor_directory}/kubectl.sh"  \
       -e "DS_KUDO_VERSION=v${DS_KUDO_VERSION}" \
       -e "TOOLING_DIRECTORY=${container_tooling_directory}" \
       -v "${kubeconfig}:${container_kubeconfig}:ro" \
       -v "${project_directory}:${container_project_directory}" \
       -v "${TOOLING_DIRECTORY}:${container_tooling_directory}" \
       -w "${container_project_directory}" \
       "${INTEGRATION_TESTS_DOCKER_IMAGE}" \
       bash -c "${container_project_directory}/shared/deploy-kudo-controller-and-crds.sh"
