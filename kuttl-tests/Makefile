KUDO_VERSION=0.12.0
KUTTL_VERSION=local
KUBERNETES_VERSION=1.18.2
KIND_VERSION=0.8.1

OS=$(shell uname -s | tr '[:upper:]' '[:lower:]')
KUDO_MACHINE=$(shell uname -m)
MACHINE=$(shell uname -m)
ifeq "$(MACHINE)" "x86_64"
  MACHINE=amd64
endif

export PATH := $(shell pwd)/bin:$(PATH)
export ROOT_DIR=$(realpath -L ./..)
export DOCKER_API_VERSION=1.39

ARTIFACTS=kuttl-dist

.PHONY: lint
lint:
	golangci-lint run

bin/kubectl_$(KUBERNETES_VERSION)_$(OS):
	mkdir -p bin/
	curl -Lo bin/kubectl_$(KUBERNETES_VERSION)_$(OS) https://storage.googleapis.com/kubernetes-release/release/v$(KUBERNETES_VERSION)/bin/$(OS)/$(MACHINE)/kubectl
	chmod +x bin/kubectl_$(KUBERNETES_VERSION)_$(OS)

bin/kubectl: bin/kubectl_$(KUBERNETES_VERSION)_$(OS)
	ln -sf ./kubectl_$(KUBERNETES_VERSION)_$(OS) ./bin/kubectl

bin/kubectl-kuttl_$(KUTTL_VERSION)_$(OS):
	mkdir -p bin/
	curl -Lo bin/kubectl-kuttl_$(KUTTL_VERSION)_$(OS) https://github.com/kudobuilder/kuttl/releases/download/v$(KUTTL_VERSION)/kubectl-kuttl_$(KUTTL_VERSION)_$(OS)_$(KUDO_MACHINE)
	chmod +x bin/kubectl-kuttl_$(KUTTL_VERSION)_$(OS)

bin/kubectl-kuttl: bin/kubectl-kuttl_$(KUTTL_VERSION)_$(OS)
	echo "Linking kubectl-kuttl_$(KUTTL_VERSION)_$(OS)"
	ln -sf ./kubectl-kuttl_$(KUTTL_VERSION)_$(OS) ./bin/kubectl-kuttl

bin/kubectl-kudo_$(KUDO_VERSION)_$(OS):
	mkdir -p bin/
	curl -Lo bin/kubectl-kudo_$(KUDO_VERSION)_$(OS) https://github.com/kudobuilder/kudo/releases/download/v$(KUDO_VERSION)/kubectl-kudo_$(KUDO_VERSION)_$(OS)_$(KUDO_MACHINE)
	chmod +x bin/kubectl-kudo_$(KUDO_VERSION)_$(OS)

bin/kubectl-kudo: bin/kubectl-kudo_$(KUDO_VERSION)_$(OS)
	ln -sf ./kubectl-kudo_$(KUDO_VERSION)_$(OS) ./bin/kubectl-kudo

bin/kind_$(KIND_VERSION)_$(OS):
	mkdir -p bin/
	curl -Lo bin/kind_$(KIND_VERSION)_$(OS) https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-$(OS)-$(MACHINE)
	chmod +x bin/kind_$(KIND_VERSION)_$(OS)

bin/kind: bin/kind_$(KIND_VERSION)_$(OS)
	ln -sf ./kind_$(KIND_VERSION)_$(OS) ./bin/kind

.PHONY: install-kuttl
install-kuttl: bin/kubectl-kuttl bin/kubectl-kudo bin/kubectl bin/kind

.PHONY: kind-test
kind-test: install-kuttl render-test-templates
	mkdir -p kuttl-dist/
	go get github.com/jstemmer/go-junit-report
	kind delete cluster
	kubectl kuttl test --config=./suites/kuttl-common.yaml --artifacts-dir=$(ARTIFACTS) 2>&1 |tee /dev/fd/2 | go-junit-report -set-exit-code > kuttl-dist/common-junit.xml
	kubectl kuttl test --config=./suites/kuttl-failure-recovery.yaml --artifacts-dir=$(ARTIFACTS) 2>&1 |tee /dev/fd/2 | go-junit-report -set-exit-code > kuttl-dist/failure-recovery-junit.xml

.PHONY: clean
clean:
	rm -rf ./bin ./kuttl-dist ./kubeconfig

.PHONY: render-test-templates
render-test-templates:
	./render-test-templates.sh
