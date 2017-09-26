# ref. http://postd.cc/auto-documented-makefile/

COVERAGE := $(shell mktemp)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

todo: ## List up TODO tasks
	@grep --color=auto -nR TODO src

test: ## Run tests only
	go test $$(glide novendor) $(OPT)

test-coverage: ## Run tests and show coverage in browser
	go test -v -coverprofile=$(COVERAGE) -covermode=count
	go tool cover -html=$(COVERAGE)
