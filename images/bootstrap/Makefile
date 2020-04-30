
.PHONY: goimports
goimports:
	goimports -d .

.PHONY: lint
lint: goimports
	golangci-lint run

test:
	go get -u gotest.tools/gotestsum
	gotestsum --junitfile bootstrap-unit-tests.xml