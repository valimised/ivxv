include ../../../common/go/gopath.mk
include ../../../common/go/govar.mk

.DEFAULT_GOAL := all

# Linting tools produces a lot of false positives because of testdata. They are
# only used to prepare the source code for building, so they are not as
# critical and linting them during testing can be skipped.
.PHONY: test
test: all
	$(GO) test -v ./...

# Do not install tools.
.PHONY: install
install:

# Remove additional files in clean.
CLEAN := src/ivxv.ee/cmd/gen/gentmpl.go

.PHONY: clean
clean:
	$(MAKE) -f ../../../common/go/common.mk clean
	rm -rf $(CLEAN)

export NOGEN := true  # See common.mk for an explanation of this.
%: force
	$(MAKE) -f ../../../common/go/common.mk $@
force ../../../common/go/gopath.mk ../../../common/go/govar.mk Makefile: ;
