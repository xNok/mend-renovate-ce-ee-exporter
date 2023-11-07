NAME          := mend-renovate-ce-ee-exporter
FILES         := $(shell git ls-files */*.go)
COVERAGE_FILE := coverage.out
REPOSITORY    := xnok/$(NAME)
.DEFAULT_GOAL := help

.PHONY: fmt
fmt: ## Format source code
	go run mvdan.cc/gofumpt@v0.5.0 -w $(shell git ls-files **/*.go)
	go run github.com/daixiang0/gci@v0.11.2 write -s standard -s default -s "prefix(github.com/xnok)" .

.PHONY: lint
lint: ## Run all lint related tests upon the codebase
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2 run -v --fast

.PHONY: test
test: ## Run the tests against the codebase
	@rm -rf $(COVERAGE_FILE)
	go test -v -count=1 -race ./... -coverprofile=$(COVERAGE_FILE)
	@go tool cover -func $(COVERAGE_FILE) | awk '/^total/ {print "coverage: " $$3}'

.PHONY: protoc
protoc: ## Generate golang from .proto files
	@command -v protoc 2>&1 >/dev/null        || (echo "protoc needs to be available in PATH: https://github.com/protocolbuffers/protobuf/releases"; false)
	@command -v protoc-gen-go 2>&1 >/dev/null || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/monitor/protobuf/monitor.proto