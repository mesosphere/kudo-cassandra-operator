#!/usr/bin/env bash

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}")"

# shellcheck source=metadata.sh
source "${project_directory}/metadata.sh"

FAILED=0

go get golang.org/x/tools/cmd/goimports
goimports -d . | awk 'BEGIN{had_data=0}{print;had_data=1}END{exit had_data}'
FAILED=${FAILED-$?}

cd "${project_directory}"
./tools/compile_templates.sh --check-only
FAILED=${FAILED-$?}

exit ${FAILED}

