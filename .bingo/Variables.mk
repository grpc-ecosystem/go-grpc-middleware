# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.2.2. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Bellow generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for bingo variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(BINGO)
#	@echo "Running bingo"
#	@$(BINGO) <flags/args..>
#
BINGO := $(GOBIN)/bingo-v0.2.2
$(BINGO): .bingo/bingo.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/bingo-v0.2.2"
	@cd .bingo && $(GO) build -modfile=bingo.mod -o=$(GOBIN)/bingo-v0.2.2 "github.com/bwplotka/bingo"

FAILLINT := $(GOBIN)/faillint-v1.5.0
$(FAILLINT): .bingo/faillint.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/faillint-v1.5.0"
	@cd .bingo && $(GO) build -modfile=faillint.mod -o=$(GOBIN)/faillint-v1.5.0 "github.com/fatih/faillint"

GOIMPORTS := $(GOBIN)/goimports-v0.0.0-20200529172331-a64b76657301
$(GOIMPORTS): .bingo/goimports.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/goimports-v0.0.0-20200529172331-a64b76657301"
	@cd .bingo && $(GO) build -modfile=goimports.mod -o=$(GOBIN)/goimports-v0.0.0-20200529172331-a64b76657301 "golang.org/x/tools/cmd/goimports"

GOLANGCI_LINT := $(GOBIN)/golangci-lint-v1.27.0
$(GOLANGCI_LINT): .bingo/golangci-lint.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/golangci-lint-v1.27.0"
	@cd .bingo && $(GO) build -modfile=golangci-lint.mod -o=$(GOBIN)/golangci-lint-v1.27.0 "github.com/golangci/golangci-lint/cmd/golangci-lint"

MISSPELL := $(GOBIN)/misspell-v0.3.4
$(MISSPELL): .bingo/misspell.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/misspell-v0.3.4"
	@cd .bingo && $(GO) build -modfile=misspell.mod -o=$(GOBIN)/misspell-v0.3.4 "github.com/client9/misspell/cmd/misspell"

PROTOC_GEN_GOGOFAST := $(GOBIN)/protoc-gen-gogofast-v1.3.1
$(PROTOC_GEN_GOGOFAST): .bingo/protoc-gen-gogofast.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/protoc-gen-gogofast-v1.3.1"
	@cd .bingo && $(GO) build -modfile=protoc-gen-gogofast.mod -o=$(GOBIN)/protoc-gen-gogofast-v1.3.1 "github.com/gogo/protobuf/protoc-gen-gogofast"

