
.PHONY: lint
lint:
	golangci-lint run

test:
	go get -u gotest.tools/gotestsum
	gotestsum --junitfile bootstrap-unit-tests.xml