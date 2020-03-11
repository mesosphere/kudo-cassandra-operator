#!/usr/bin/env bash

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}")"

# shellcheck source=metadata.sh
source "${project_directory}/metadata.sh"

FAILED=0

cd "${project_directory}"
./tools/compile_templates.sh --check-only
FAILED=${FAILED-$?}

exit ${FAILED}

