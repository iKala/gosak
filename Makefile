.PHONY: asset deps install test
.DEFAULT_GOAL := help

install: ## Install the app
	go install ./...

test: ## Run all test
	go test -v ./...

deps: ## Install all dependencies
	# below is for test
	go get github.com/stretchr/testify/assert

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
