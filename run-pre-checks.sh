#!/usr/bin/env bash

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}")"

# We need to ignore the suffix for the purpose of checking templates.
IMAGE_DISAMBIGUATION_SUFFIX="" "${project_directory}/tools/compile_templates.sh" --check-only
