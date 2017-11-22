SHELL="/bin/bash"

GOFILES_NOVENDOR = $(shell go list ./... | grep -v /vendor/)

default:


docs:
	./scripts/fixup.sh

validate:
	go fmt $(GOFILES_NOVENDOR)
	go vet $(GOFILES_NOVENDOR)

test: validate
	./scripts/test_all.sh

.PHONY: default docs validate test
