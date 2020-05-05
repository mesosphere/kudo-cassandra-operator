#!/usr/bin/env bash
# shellcheck disable=SC2039

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(realpath -L $script_directory)"

# shellcheck source=metadata.sh
source "${project_directory}/metadata.sh"

if [[ -n ${IMAGE_DISAMBIGUATION_SUFFIX:-} ]]; then
  # Refresh templated files to pick up the suffix, if explicitly set.
  "${project_directory}/tools/compile_templates.sh"
fi

readonly kubeconfig="${KUBECONFIG:-${HOME}/.kube/config}"
readonly artifacts_directory="${DS_TEST_ARTIFACTS_DIRECTORY:-${project_directory}/artifacts}"

mkdir -p "${artifacts_directory}"

readonly container_kubeconfig="/root/.kube/config"
readonly container_project_directory="/${PROJECT_NAME}"
readonly container_operator_directory="${container_project_directory}/operator"
readonly container_vendor_directory="${container_project_directory}/shared/vendor"
readonly container_artifacts_directory="${container_project_directory}/artifacts"

# Note: DS_KUDO_VERSION is used by the shared data-services-kudo tooling.

docker run \
       --rm \
       -e "KUBECONFIG=${container_kubeconfig}" \
       -e "KUBECTL_PATH=${container_vendor_directory}/kubectl.sh"  \
       -e "DS_KUDO_VERSION=v${KUDO_VERSION}" \
       -e "OPERATOR_DIRECTORY=${container_operator_directory}" \
       -e "VENDOR_DIRECTORY=${container_vendor_directory}" \
       -e "TEST_ARTIFACTS_DIRECTORY=${container_artifacts_directory}" \
       -e "AWS_ACCESS_KEY_ID" \
       -e "AWS_SECRET_ACCESS_KEY" \
       -e "BUILD_NUMBER" \
       -e "TEAMCITY_PROJECT_NAME" \
       -v "${kubeconfig}:${container_kubeconfig}:ro" \
       -v "${OPERATOR_DIRECTORY}:${container_operator_directory}" \
       -v "${project_directory}:${container_project_directory}" \
       -v "${VENDOR_DIRECTORY}:${container_vendor_directory}" \
       -v "${artifacts_directory}:${container_artifacts_directory}" \
       -w "${container_project_directory}" \
       "${INTEGRATION_TESTS_DOCKER_IMAGE}" \
       bash -c "${container_project_directory}/tests/run.sh"
