include ../../../common/go/gopath.mk
include ../../../common/go/govar.mk

.DEFAULT_GOAL := all

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
