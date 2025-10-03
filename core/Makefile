# TIVI Core Go project absolute path dir
ROOTDIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# golangci-lint version
GOLANGCI_LINT_VERSION := 1.62.0

# Go test packages to be linted, '/...' denotes "recursive searching"
GO_TEST_PACKAGES := $(ROOTDIR)math/... $(ROOTDIR)crypto/ecdsa/... $(ROOTDIR)crypto/elgamal/... $(ROOTDIR)crypto/x509/...

# Copy everything from that file here
include $(ROOTDIR)govar.mk

.PHONY: help
help:
	@echo "usage: make help       Show this help message"
	@echo "       make lint       Run code linter"
	@echo "       make test       Run unit tests"
	@echo "       make codegen    Auto-generate code"
	@echo "       make clean      Clean up auto-generated code"

# Lint Go code inside a Docker
.PHONY: lint
lint: codegen
	docker run --rm -v $(ROOTDIR):/linter -w /linter golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION) golangci-lint run

# Test only specific Go packages
.PHONY: test
test: lint
	go test -v $(GO_TEST_PACKAGES)

# Target prefixed with "--" are not seen to the call `make [target]`
.PHONY: --build-errorgen
--build-errorgen: clean
	$(MAKE) -C cmd/errorgen all

# Generate Go code that has a '//go:generate' directive.
# Generate Go code custom error structs on fly
.PHONY: codegen
codegen: --build-errorgen
	$(ROOTDIR)cmd/errorgen/bin/errorgen -base $(ROOTDIR) ./...

# Clean all Go binaries and auto-generated files
.PHONY: clean
clean:
	find . -not -path \*/testdata/\* \( \
		-name gen_types.go -o \
		-name gen_types_test.go -o \
		-name gen_import.go -o \
		-name gen_import_dev.go \
		\) -delete
	 -$(GO) clean -i ./...
	rm -rf bin/
