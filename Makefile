include .bingo/Variables.mk
SHELL=/bin/bash

PROVIDER_MODULES ?= $(shell ls -d $(PWD)/providers/*)
MODULES          ?= $(PROVIDER_MODULES) $(PWD)/

GOBIN             ?= $(firstword $(subst :, ,${GOPATH}))/bin

GO111MODULE       ?= on
export GO111MODULE
GOPROXY           ?= https://proxy.golang.org
export GOPROXY

.PHONY: fmt
fmt:
	@echo "Running fmt for all modules: ${MODULES}"
	@$(GOIMPORTS) -local github.com/grpc-ecosystem/go-grpc-middleware/v2 -w ${MODULES}

.PHONY: proto
proto: ## Generates Go files from Thanos proto files.
proto: $(GOIMPORTS) $(PROTOC) $(PROTOC_GEN_GOGOFAST) ./grpctesting/testpb/test.proto
	@GOIMPORTS_BIN="$(GOIMPORTS)" PROTOC_BIN="$(PROTOC)" PROTOC_GEN_GOGOFAST_BIN="$(PROTOC_GEN_GOGOFAST)" scripts/genproto.sh

.PHONY: test
test:
	@echo "Running tests for all modules: ${MODULES}"
	@$(foreach dir,$(PROVIDER_MODULES),$(MAKE) test_module DIR=$(dir))
	./scripts/test_all.sh

.PHONY: test_module
test_module:
	cd $(DIR) && go test -v -race ./...

.PHONY: lint
lint: fmt
	@echo "Running lint for all modules: ${MODULES}"
	@$(foreach dir,$(MODULES),$(MAKE) lint_module DIR=$(dir))

.PHONY: lint_module
# PROTIP:
# Add
#      --cpu-profile-path string   Path to CPU profile output file
#      --mem-profile-path string   Path to memory profile output file
# to debug big allocations during linting.
lint: ## Runs various static analysis against our code.
lint_module: $(FAILLINT) $(GOLANGCI_LINT) $(MISSPELL)
	$(call require_clean_work_tree,"detected not clean master before running lint")
	@echo ">> verifying modules being imported"
	@$(FAILLINT) -paths "errors=github.com/pkg/errors,fmt.{Print,Printf,Println}" ./...
	@echo ">> examining all of the Go files"
	@go vet -stdmethods=false ./...
	@echo ">> linting all of the Go files GOGC=${GOGC}"
	@$(GOLANGCI_LINT) run
	@echo ">> detecting misspells"
	@find . -type f | grep -v vendor/ | grep -vE '\./\..*' | xargs $(MISSPELL) -error
	@echo ">> detecting white noise"
	@find . -type f \( -name "*.md" -o -name "*.go" \) | SED_BIN="$(SED)" xargs scripts/cleanup-white-noise.sh
	@echo ">> ensuring generated proto files are up to date"
	@$(MAKE) proto
	$(call require_clean_work_tree,"detected white noise or/and files without copyright; run 'make lint' file and commit changes.")

$(PROTOC):
	@mkdir -p $(TMP_GOPATH)
	@echo ">> fetching protoc@${PROTOC_VERSION}"
	@PROTOC_VERSION="$(PROTOC_VERSION)" TMP_GOPATH="$(TMP_GOPATH)" scripts/installprotoc.sh
	@echo ">> installing protoc@${PROTOC_VERSION}"
	@mv -- "$(TMP_GOPATH)/bin/protoc" "$(GOBIN)/protoc-$(PROTOC_VERSION)"
	@echo ">> produced $(GOBIN)/protoc-$(PROTOC_VERSION)"