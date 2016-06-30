.PHONY: asset deps restore install test
.DEFAULT_GOAL := help

deps: ## Pack the app dependencies
	rm -rf Godeps/
	rm -rf vendor/
	godep save github.com/iKala/gosak

restore: ## Restore the app dependencies
	go get github.com/tools/godep
	godep restore

install: deps ## Install
	go install ./...

test: ## Run all test
	go test -v ./...

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
