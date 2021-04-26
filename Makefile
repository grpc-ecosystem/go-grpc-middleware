include .bingo/Variables.mk

SHELL=/bin/bash

PROVIDER_MODULES ?= $(shell find $(PWD)/providers/  -name "go.mod" | grep -v ".bingo" | xargs dirname)
MODULES          ?= $(PROVIDER_MODULES) $(PWD)/

GOBIN             ?= $(firstword $(subst :, ,${GOPATH}))/bin

TMP_GOPATH        ?= /tmp/gopath

GO111MODULE       ?= on
export GO111MODULE
GOPROXY           ?= https://proxy.golang.org
export GOPROXY

define require_clean_work_tree
	@git update-index -q --ignore-submodules --refresh

    @if ! git diff-files --quiet --ignore-submodules --; then \
        echo >&2 "cannot $1: you have unstaged changes."; \
        git diff-files --name-status -r --ignore-submodules -- >&2; \
        echo >&2 "Please commit or stash them."; \
        exit 1; \
    fi

    @if ! git diff-index --cached --quiet HEAD --ignore-submodules --; then \
        echo >&2 "cannot $1: your index contains uncommitted changes."; \
        git diff-index --cached --name-status -r --ignore-submodules HEAD -- >&2; \
        echo >&2 "Please commit or stash them."; \
        exit 1; \
    fi

endef

all: fmt proto lint test

.PHONY: fmt
fmt: $(GOIMPORTS)
	@echo "Running fmt for all modules: $(MODULES)"
	@$(GOIMPORTS) -local github.com/grpc-ecosystem/go-grpc-middleware/v2 -w $(MODULES)

.PHONY: test
test:
	@echo "Running tests for all modules: $(MODULES)"
	for dir in $(MODULES) ; do \
		$(MAKE) test_module DIR=$${dir} ; \
	done
	@./scripts/test_all.sh

.PHONY: test_module
test_module:
	@echo "Running tests for dir: $(DIR)"
	cd $(DIR) && go test -v -race ./...

.PHONY: deps
deps:
	@echo "Running deps tidy for all modules: $(MODULES)"
	for dir in $(MODULES) ; do \
  		echo "$${dir}"; \
		cd $${dir} && go mod tidy; \
	done

.PHONY: lint
# PROTIP:
# Add
#      --cpu-profile-path string   Path to CPU profile output file
#      --mem-profile-path string   Path to memory profile output file
# to debug big allocations during linting.
lint: ## Runs various static analysis tools against our code.
lint: $(BUF) $(COPYRIGHT) fmt
	@echo ">> lint proto files"
	@$(BUF) lint
	@echo "Running lint for all modules: $(MODULES)"
	./scripts/git-tree.sh
	for dir in $(MODULES) ; do \
		$(MAKE) lint_module DIR=$${dir} ; \
	done
	@$(call require_clean_work_tree,"lint and format files")

	@echo ">> ensuring copyright headers"
	@$(COPYRIGHT) $(shell go list -f "{{.Dir}}" ./... | xargs -i find "{}" -name "*.go")
	@$(call require_clean_work_tree,"set copyright headers")
	@echo ">> ensured all .go files have copyright headers"

.PHONY: lint_module
# PROTIP:
# Add
#      --cpu-profile-path string   Path to CPU profile output file
#      --mem-profile-path string   Path to memory profile output file
# to debug big allocations during linting.
lint_module: ## Runs various static analysis against our code.
lint_module: $(FAILLINT) $(GOLANGCI_LINT) $(MISSPELL)
	@echo ">> verifying modules being imported"
	@cd $(DIR) && $(FAILLINT) -paths "errors=github.com/pkg/errors,fmt.{Print,Printf,Println}" ./...
	@echo ">> examining all of the Go files"
	@cd $(DIR) && go vet -stdmethods=false ./...
	@echo ">> linting all of the Go files GOGC=${GOGC}"
	@cd $(DIR) && $(GOLANGCI_LINT) run
	@./scripts/git-tree.sh


# For protoc naming matters.
PROTOC_GEN_GO_CURRENT := $(TMP_GOPATH)/protoc-gen-go
PROTOC_GEN_GO_GRPC_CURRENT := $(TMP_GOPATH)/protoc-gen-go-grpc
PROTO_TEST_DIR := testing/testpb/v1

.PHONY: proto
proto: ## Generate testing protobufs
proto: $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) $(PROTO_TEST_DIR)/test.proto
	@mkdir -p $(TMP_GOPATH)
	@cp $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_CURRENT)
	@cp $(PROTOC_GEN_GO_GRPC) $(PROTOC_GEN_GO_GRPC_CURRENT)
	@echo ">> generating $(PROTO_TEST_DIR)"
	@PATH=$(GOBIN):$(TMP_GOPATH) $(BUF) protoc \
		-I $(PROTO_TEST_DIR) \
		--go_out=$(PROTO_TEST_DIR)/../ \
		--go-grpc_out=$(PROTO_TEST_DIR)/../ \
	    $(PROTO_TEST_DIR)/*.proto

    