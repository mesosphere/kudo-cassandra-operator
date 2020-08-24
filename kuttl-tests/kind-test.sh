#!/usr/bin/env bash

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}/..")"

# shellcheck source=../metadata.sh
source "${project_directory}/metadata.sh"

KUBECTL_VERSION=1.18.4
KUTTL_VERSION=0.5.0
KUBECTL_KUDO_VERSION=${DS_KUDO_VERSION#v}
KIND_VERSION=0.8.1

ARTIFACTS=kuttl-dist

OS=$(shell uname -s | tr '[:upper:]' '[:lower:]')
KUDO_MACHINE=$(shell uname -m)
MACHINE=$(shell uname -m)
ifeq "$(MACHINE)" "x86_64"
  MACHINE=amd64
endif

mkdir -p bin/

curl -Lo "bin/kubectl_${KUBECTL_VERSION}_${OS}" "https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/${OS}/${MACHINE}/kubectl"
chmod +x "bin/kubectl_${KUBECTL_VERSION}_${OS}"
ln -sf "bin/kubectl_${KUBECTL_VERSION}_${OS}" bin/kubectl

curl -Lo "bin/kubectl-kuttl_${KUTTL_VERSION}_${OS}" "https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_${OS}_${KUDO_MACHINE}"
chmod +x "bin/kubectl-kuttl_${KUTTL_VERSION}_${OS}"
ln -sf "bin/kubectl-kuttl_${KUTTL_VERSION}_${OS}" bin/kubectl-kuttl

curl -Lo "bin/kubectl-kudo_${KUBECTL_KUDO_VERSION}_${OS}" "https://github.com/kudobuilder/kudo/releases/download/v${KUBECTL_KUDO_VERSION}/kubectl-kudo_${KUBECTL_KUDO_VERSION}_${OS}_${KUDO_MACHINE}"
chmod +x "bin/kubectl-kudo_${KUBECTL_KUDO_VERSION}_${OS}"
ln -sf "bin/kubectl-kudo_${KUBECTL_KUDO_VERSION}_${OS}" bin/kubectl-kudo

curl -Lo "bin/kind_${KIND_VERSION}_${OS}" "https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-${OS}-${MACHINE}"
chmod +x "bin/kind_${KIND_VERSION}_${OS}"
ln -sf "bin/kind_${KIND_VERSION}_${OS}" bin/kind

mkdir -p $ARTIFACTS
go get github.com/jstemmer/go-junit-report
kubectl kuttl test --config=./suites/kuttl-common.yaml --artifacts-dir=${ARTIFACTS} 2>&1 | tee /dev/fd/2 | go-junit-report -set-exit-code > kuttl-dist/common-junit.xml
kubectl kuttl test --config=./suites/kuttl-failure-recovery.yaml --artifacts-dir=${ARTIFACTS} 2>&1 | tee /dev/fd/2 | go-junit-report -set-exit-code > kuttl-dist/failure-recovery-junit.xml
