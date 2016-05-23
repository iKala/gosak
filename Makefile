.PHONY: asset deps install test
.DEFAULT_GOAL := help

install: ## Install
	go install ./...

test: ## Run all test
	go test -v ./...

deps: ## Install all dependencies
	go get github.com/bluele/slack
	# below is for test
	go get github.com/stretchr/testify/assert

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
