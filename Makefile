include .bingo/Variables.mk

SHELL=/bin/bash

PROVIDER_MODULES ?= $(shell find $(PWD)/providers/  -name "go.mod" | grep -v ".bingo" | xargs dirname)
MODULES          ?= $(PROVIDER_MODULES) $(PWD)/

GOBIN             ?= $(firstword $(subst :, ,${GOPATH}))/bin

// TODO(bwplotka): Move to buf.
PROTOC_VERSION    ?= 3.12.3
PROTOC            ?= $(GOBIN)/protoc-$(PROTOC_VERSION)
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

.PHONY: proto
proto: ## Generates Go files from Thanos proto files.
proto: $(GOIMPORTS) $(PROTOC) $(PROTOC_GEN_GOGOFAST) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) ./grpctesting/testpb/test.proto
	@GOIMPORTS_BIN="$(GOIMPORTS)" PROTOC_BIN="$(PROTOC)" PROTOC_GEN_GO_BIN="$(PROTOC_GEN_GO)" PROTOC_GEN_GO_GRPC_BIN="$(PROTOC_GEN_GO_GRPC)" PROTOC_GEN_GOGOFAST_BIN="$(PROTOC_GEN_GOGOFAST)" scripts/genproto.sh

.PHONY: test
test:
	@echo "Running tests for all modules: $(MODULES)"
	for dir in $(MODULES) ; do \
		$(MAKE) test_module DIR=$${dir} ; \
	done
	./scripts/test_all.sh

.PHONY: test_module
test_module:
	@echo "Running tests for dir: $(DIR)"
	cd $(DIR) && go test -v -race ./...

.PHONY: deps
deps:
	@echo "Running deps tidy for all modules: $(MODULES)"
	for dir in $(MODULES) ; do \
		cd $${dir} && go mod tidy; \
	done

.PHONY: lint
# PROTIP:
# Add
#      --cpu-profile-path string   Path to CPU profile output file
#      --mem-profile-path string   Path to memory profile output file
# to debug big allocations during linting.
lint: ## Runs various static analysis tools against our code.
lint: fmt proto
	@echo "Running lint for all modules: $(MODULES)"
	./scripts/git-tree.sh
	for dir in $(MODULES) ; do \
		$(MAKE) lint_module DIR=$${dir} ; \
	done

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

# TODO(bwplotka): Move to buf.
$(PROTOC):
	@mkdir -p $(TMP_GOPATH)
	@echo ">> fetching protoc@${PROTOC_VERSION}"
	@PROTOC_VERSION="$(PROTOC_VERSION)" TMP_GOPATH="$(TMP_GOPATH)" scripts/installprotoc.sh
	@echo ">> installing protoc@${PROTOC_VERSION}"
	@mv -- "$(TMP_GOPATH)/bin/protoc" "$(GOBIN)/protoc-$(PROTOC_VERSION)"
	@echo ">> produced $(GOBIN)/protoc-$(PROTOC_VERSION)"
