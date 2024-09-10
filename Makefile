install: ## Install all dependencies
	go get ./...

update: ## Update dependencies
	go get -u ./... && go mod tidy && go mod vendor

fmt: ## Run go fmt against code
	go fmt ./pkg/...

vet: ## Run go vet against code
	go vet ./pkg/... 

build: ## Build the web-gateway service
	go build -o blockmandu ./cmd/

tools: ## installs analysis tools
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

analyse: ## runs go tools analysis tools
	fieldalignment -fix ./...

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: install update fmt vet build tools analyse
