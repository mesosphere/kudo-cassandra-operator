#!/usr/bin/env bash

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}/..")"

# shellcheck source=../metadata.sh
source "${project_directory}/metadata.sh"

# Note: the following environment variables are required by the shared
# data-services-kudo tooling.
export DS_KUDO_VERSION="${DS_KUDO_VERSION:-v${KUDO_VERSION}}"
export KUBECONFIG="${KUBECONFIG:-${HOME}/.kube/config}"
export KUBECTL_PATH="${KUBECTL_PATH:-${VENDOR_DIRECTORY}/kubectl.sh}"
export OPERATOR_DIRECTORY="${OPERATOR_DIRECTORY:-${project_directory}/operator}"
export VENDOR_DIRECTORY="${VENDOR_DIRECTORY:-${project_directory}/shared/vendor}"

# Tests that take longer than this in seconds are marked as slow
export GINKO_SLOW_TEST_THRESHOLD=240

if [ -n "${GOPATH:-}" ]; then
  export GINKGO_PATH="${GOPATH}/bin/ginkgo"
else
  export GINKGO_PATH="${HOME}/go/bin/ginkgo"
fi

# https://github.com/golang/go/wiki/Modules.
# FIXME(mpereira): is this necessary? Leaving it commented for now.
# export GO111MODULE=on

# Give more priority to vendored executables.
export PATH=${VENDOR_DIRECTORY}:${PATH}

#cd "${project_directory}"
#./tools/compile_templates.sh --check-only

cd "${project_directory}/tests"

go mod edit -require "github.com/kudobuilder/kudo@${DS_KUDO_VERSION}"
go install github.com/onsi/ginkgo/ginkgo

${KUBECTL_PATH} kudo version

if [[ -z ${1-} ]]
then
    ${GINKGO_PATH} --slowSpecThreshold=${GINKO_SLOW_TEST_THRESHOLD:-240} --succinct=false -v ./suites/... ${TESTS_FOCUS:+--ginkgo.focus=${TESTS_FOCUS}}
else
    ${GINKGO_PATH} --slowSpecThreshold=${GINKO_SLOW_TEST_THRESHOLD:-240} --succinct=false -v ./suites/$1/... ${TESTS_FOCUS:+--ginkgo.focus=${TESTS_FOCUS}}
fi
