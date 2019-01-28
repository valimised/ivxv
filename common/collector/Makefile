include ../go/gopath.mk

.DEFAULT_GOAL := all

.PHONY: install
install: all
	$(MAKE) -C config $@
	$(MAKE) -C scripts $@
	$(MAKE) -f ../go/common.mk $@

# Remove additional files in clean.
CLEAN := src/ivxv.ee/server/tlsciphersuites.go

.PHONY: clean
clean:
	$(MAKE) -f ../go/common.mk clean
	rm -rf $(CLEAN)

%: force
	$(MAKE) -f ../go/common.mk $@
force ../go/gopath.mk Makefile: ;
