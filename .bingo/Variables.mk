# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.8. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
BINGO_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Below generated variables ensure that every time a tool under each variable is invoked, the correct version
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
BINGO := $(GOBIN)/bingo-v0.3.0
$(BINGO): $(BINGO_DIR)/bingo.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/bingo-v0.3.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=bingo.mod -o=$(GOBIN)/bingo-v0.3.0 "github.com/bwplotka/bingo"

BUF := $(GOBIN)/buf-v1.26.1
$(BUF): $(BINGO_DIR)/buf.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/buf-v1.26.1"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=buf.mod -o=$(GOBIN)/buf-v1.26.1 "github.com/bufbuild/buf/cmd/buf"

COPYRIGHT := $(GOBIN)/copyright-v0.0.0-20210326193628-425a09c04e05
$(COPYRIGHT): $(BINGO_DIR)/copyright.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/copyright-v0.0.0-20210326193628-425a09c04e05"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=copyright.mod -o=$(GOBIN)/copyright-v0.0.0-20210326193628-425a09c04e05 "github.com/efficientgo/tools/copyright"

FAILLINT := $(GOBIN)/faillint-v1.11.0
$(FAILLINT): $(BINGO_DIR)/faillint.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/faillint-v1.11.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=faillint.mod -o=$(GOBIN)/faillint-v1.11.0 "github.com/fatih/faillint"

GOIMPORTS := $(GOBIN)/goimports-v0.7.0
$(GOIMPORTS): $(BINGO_DIR)/goimports.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/goimports-v0.7.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=goimports.mod -o=$(GOBIN)/goimports-v0.7.0 "golang.org/x/tools/cmd/goimports"

GOLANGCI_LINT := $(GOBIN)/golangci-lint-v1.52.0
$(GOLANGCI_LINT): $(BINGO_DIR)/golangci-lint.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/golangci-lint-v1.52.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=golangci-lint.mod -o=$(GOBIN)/golangci-lint-v1.52.0 "github.com/golangci/golangci-lint/cmd/golangci-lint"

MDOX := $(GOBIN)/mdox-v0.9.0
$(MDOX): $(BINGO_DIR)/mdox.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/mdox-v0.9.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=mdox.mod -o=$(GOBIN)/mdox-v0.9.0 "github.com/bwplotka/mdox"

MISSPELL := $(GOBIN)/misspell-v0.3.4
$(MISSPELL): $(BINGO_DIR)/misspell.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/misspell-v0.3.4"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=misspell.mod -o=$(GOBIN)/misspell-v0.3.4 "github.com/client9/misspell/cmd/misspell"

PROTOC_GEN_GO_GRPC := $(GOBIN)/protoc-gen-go-grpc-v0.0.0-20200723182653-9106c3fff523
$(PROTOC_GEN_GO_GRPC): $(BINGO_DIR)/protoc-gen-go-grpc.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/protoc-gen-go-grpc-v0.0.0-20200723182653-9106c3fff523"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=protoc-gen-go-grpc.mod -o=$(GOBIN)/protoc-gen-go-grpc-v0.0.0-20200723182653-9106c3fff523 "google.golang.org/grpc/cmd/protoc-gen-go-grpc"

PROTOC_GEN_GO := $(GOBIN)/protoc-gen-go-v1.30.0
$(PROTOC_GEN_GO): $(BINGO_DIR)/protoc-gen-go.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/protoc-gen-go-v1.30.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=protoc-gen-go.mod -o=$(GOBIN)/protoc-gen-go-v1.30.0 "google.golang.org/protobuf/cmd/protoc-gen-go"

