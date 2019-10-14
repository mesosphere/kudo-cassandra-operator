#!/usr/bin/env bash

# Depends on the following environment variables:
# - DS_KUDO_VERSION
# - KUBECONFIG
# - KUBECTL_PATH
# - OPERATOR_DIRECTORY
# - VENDOR_DIRECTORY

set -euxo pipefail

readonly SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly PROJECT_DIRECTORY="$(readlink -f "${SCRIPT_DIRECTORY}/..")"

# shellcheck source=../metadata.sh
source "${PROJECT_DIRECTORY}/metadata.sh"

# TODO(mpereira): complain if KUBECONFIG doesn't exist.
export KUBECONFIG="${KUBECONFIG:-${HOME}/.kube/config}"
export OPERATOR_DIRECTORY="${OPERATOR_DIRECTORY:-${PROJECT_DIRECTORY}/operator}"
export VENDOR_DIRECTORY="${VENDOR_DIRECTORY:-${PROJECT_DIRECTORY}/shared/vendor}"
export KUBECTL_PATH="${KUBECTL_PATH:-${VENDOR_DIRECTORY}/kubectl.sh}"
export DS_KUDO_VERSION="${DS_KUDO_VERSION:-v${KUDO_VERSION}}"

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

cd "${PROJECT_DIRECTORY}/tests"

go mod edit -require "github.com/kudobuilder/kudo@${DS_KUDO_VERSION}"
go install github.com/onsi/ginkgo/ginkgo

${KUBECTL_PATH} kudo version

${GINKGO_PATH} ./suites/... ${TESTS_FOCUS:+--ginkgo.focus=${TESTS_FOCUS}}
