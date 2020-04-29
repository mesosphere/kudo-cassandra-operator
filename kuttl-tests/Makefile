
.PHONY: lint
lint:
	golangci-lint run

KUDO_VERSION=0.12.0
KUTTL_VERSION=0.1.0
KUBERNETES_VERSION=1.17.2

OS=$(shell uname -s | tr '[:upper:]' '[:lower:]')
KUDO_MACHINE=$(shell uname -m)
MACHINE=$(shell uname -m)
ifeq "$(MACHINE)" "x86_64"
  MACHINE=amd64
endif

export PATH := $(shell pwd)/bin:$(PATH)

ARTIFACTS=dist

bin/:
	mkdir -p bin/

dist/:
	mkdir -p dist/

bin/kubectl_$(KUBERNETES_VERSION): bin/
	curl -Lo bin/kubectl_$(KUBERNETES_VERSION) https://storage.googleapis.com/kubernetes-release/release/v$(KUBERNETES_VERSION)/bin/$(OS)/$(MACHINE)/kubectl
	chmod +x bin/kubectl_$(KUBERNETES_VERSION)
	ln -sf ./kubectl_$(KUBERNETES_VERSION) ./bin/kubectl

bin/kubectl-kuttl_$(KUTTL_VERSION): bin/
	curl -Lo bin/kubectl-kuttl_$(KUTTL_VERSION) https://github.com/kudobuilder/kuttl/releases/download/v$(KUTTL_VERSION)/kubectl-kuttl_$(KUTTL_VERSION)_$(OS)_$(MACHINE)
	chmod +x bin/kubectl-kuttl_$(KUTTL_VERSION)
	ln -sf ./kubectl-kuttl_$(KUTTL_VERSION) ./bin/kubectl-kuttl

bin/kubectl-kudo_$(KUDO_VERSION): bin/
	curl -Lo bin/kubectl-kudo_$(KUDO_VERSION) https://github.com/kudobuilder/kudo/releases/download/v$(KUDO_VERSION)/kubectl-kudo_$(KUDO_VERSION)_$(OS)_$(KUDO_MACHINE)
	chmod +x bin/kubectl-kudo_$(KUDO_VERSION)
	ln -sf ./kubectl-kudo_$(KUDO_VERSION) ./bin/kubectl-kudo

.PHONY: install-kuttl
install-kuttl: bin/kubectl-kuttl_$(KUTTL_VERSION) bin/kubectl-kudo_$(KUDO_VERSION) bin/kubectl_$(KUBERNETES_VERSION)

.PHONY: test
kind-test:  dist/ install-kuttl
	go get github.com/jstemmer/go-junit-report
	which kubectl
	kubectl kuttl test --config=./kuttl-test.yaml --kind-config=./kind/kubernetes-$(KUBERNETES_VERSION).yaml --artifacts-dir=$(ARTIFACTS) 2>&1 |tee /dev/fd/2 | go-junit-report -set-exit-code > dist/addons_test_report.xml

clean:
	rm -rf ./bin ./dist ./kubeconfig
