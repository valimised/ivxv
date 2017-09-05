include ../go/gopath.mk

.DEFAULT_GOAL := all

.PHONY: install
install: all
	$(MAKE) -C config $@
	$(MAKE) -C scripts $@
	$(MAKE) -f ../go/common.mk $@

%: force
	$(MAKE) -f ../go/common.mk $@
force ../go/gopath.mk Makefile: ;
