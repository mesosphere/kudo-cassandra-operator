#!/usr/bin/env bash

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(realpath -L "${script_directory}")"
readonly artifacts_directory="${DS_TEST_ARTIFACTS_DIRECTORY:-${project_directory}/kuttl-tests}"

mkdir -p "${artifacts_directory}"

# We need to ignore the suffix for the purpose of checking templates.
IMAGE_DISAMBIGUATION_SUFFIX="" "${project_directory}/tools/compile_templates.sh" --check-only

"${project_directory}/tools/docker.sh" "${project_directory}/tools/generate_parameters_markdown.py"

"${project_directory}/tools/docker.sh" "${project_directory}/tools/format_files.sh"

source "${project_directory}/metadata.sh"
cd "${project_directory}"
set +x
if [ -n "$(git status --porcelain)" ]; then
  echo "Changes found after running one of the previous steps." >&2
  echo "Please make sure you follow the instructions in .github/pull_request_template.md" >&2
  echo "before sending a pull request." >&2
  git status --porcelain
  git diff
  exit 1
fi

# run unit tests for bootstrap binary
docker run \
  --rm \
  -v "${project_directory}:${project_directory}" \
  -w "${project_directory}"/images/bootstrap \
  "${INTEGRATION_TESTS_DOCKER_IMAGE}" \
  bash -c "make test"

mkdir -p "${artifacts_directory}"/kuttl-dist

echo "Saving KUTTL artifacts to ${artifacts_directory}/kuttl-dist"
touch ${artifacts_directory}/kuttl-dist/marker

ls -la ${artifacts_directory}/kuttl-dist/

# run KUTTL tests in ./kuttl-tests directory
docker run \
  --rm \
  -v "${project_directory}:${project_directory}" \
  -v "${artifacts_directory}/kuttl-dist:/kuttl-tests/kuttl-dist" \
  -w "${project_directory}"/kuttl-tests \
  --env-file <(env | grep BUILD_VCS_NUMBER_) \
  --privileged --network host -v /var/run/docker.sock:/var/run/docker.sock \
  "${INTEGRATION_TESTS_DOCKER_IMAGE}" \
  bash -c "make kind-test"