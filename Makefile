SHELL=/bin/bash

GODIRS_NOVENDOR   ?= $(shell go list ./... | grep -v /vendor/)
GOFILES_NOVENDOR  ?= $(shell find . -path ./vendor -prune -o -name '*.go' -print)

GOPATH            ?= $(shell go env GOPATH)
GOBIN             ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO111MODULE       ?= on
export GO111MODULE
GOPROXY           ?= https://proxy.golang.org
export GOPROXY

# TODO(bwplotka): Auto install using separate go.mod for tools.
GOIMPORTS         ?= $(GOBIN)/goimports

all: vet fmt test

fmt:
	go fmt $(GODIRS_NOVENDOR)
	@$(GOIMPORTS) -local github.com/grpc-ecosystem/go-grpc-middleware/v2 -w ${GOFILES_NOVENDOR}

vet:
	# Do not check lostcancel, they are intentional.
	# TODO(bwplotka): Explain why intentional.
	go vet -lostcancel=false $(GODIRS_NOVENDOR)

proto: ./grpctesting/testpb/test.proto
	@scripts/genproto.sh

# TODO(bwplotka): This depends on test_proto, but CI does not have it, so let's skip it for now.
test: vet
	./scripts/test_all.sh

.PHONY: all test
