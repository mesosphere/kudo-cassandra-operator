.PHONY: lint
lint:
	golangci-lint run

.PHONY: kind-test
kind-test:
	./kind-test.sh

.PHONY: clean
clean:
	rm -rf ./bin ./kuttl-dist ./kubeconfig

.PHONY: render-test-templates
render-test-templates:
	./render-test-templates.sh
