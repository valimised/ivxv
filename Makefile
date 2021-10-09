include common/go/gopath.mk

JAVADIRS := common/java key processor auditor
GODIRS   := common/tools/go common/collector proxy dds mid choices voting verification storage

DIRS        := $(JAVADIRS) $(GODIRS) systemd Documentation
TESTDIRS    := $(patsubst %,test-%,$(JAVADIRS) $(GODIRS))
INSTALLDIRS := $(patsubst %,install-%,$(JAVADIRS) $(GODIRS) systemd)
CLEANDIRS   := $(patsubst %,clean-%,$(DIRS))

export ROOT_BUILD=true

# Needed to include newlines in $(forach) loops. Must contain two empty lines.
define NEWLINE


endef

.PHONY: help
help:
	@echo "usage: make all           Build all components"
	@echo "       make test          Run unit tests for all components"
	@echo "       make install       Install all collector components to \$$DESTDIR"
	@echo "       make install-java  Install all application components to \$$DESTDIR"
	@echo "       make clean         Clean the repository"
	@echo
	@echo "       make <component>          Build the component (listed below)"
	@echo "       make test-<component>     Run unit tests for the component"
	@echo "       make install-<component>  Install the component to \$$DESTDIR"
	@echo "       make clean-<component>    Clean the component"
	@echo
	@echo "       make external         Checkout common/external to the expected version"
	@echo "       make update-external  Checkout common/external to the latest version"
	@echo "       make gopath           Print the IVXV GOPATH to standard output"
	@echo "       make version          Update version numbers in all known locations to"
	@echo "                             the last entry in debian/changelog"
	@echo
	@echo "Components:"
	$(foreach component,$(filter-out $(JAVADIRS) $(GODIRS),$(DIRS)),@echo "  $(component)"$(NEWLINE))
	@echo
	@echo "  java  (meta-component which includes all of the following)"
	$(foreach component,$(JAVADIRS),@echo "  $(component)"$(NEWLINE))
	@echo
	@echo "  go  (meta-component which includes all of the following)"
	$(foreach component,$(GODIRS),@echo "  $(component)"$(NEWLINE))
	@echo
	@echo "All rules can be suffixed with \"-dev\" to call them in development mode instead,"
	@echo "e.g., make all-dev."
	@echo
	@echo "Read README.rst in both the repository root and specific component directories"
	@echo "for more details."

.PHONY: all
all: $(DIRS)

.PHONY: java
java: $(JAVADIRS)

.PHONY: go
go: $(GODIRS)

.PHONY: $(DIRS)
$(DIRS):
	$(MAKE) -C $@

.PHONY: test
test: $(TESTDIRS)

.PHONY: test-java
test-java: $(JAVADIRS:%=test-%)

.PHONY: test-go
test-go: $(GODIRS:%=test-%)

.PHONY: test-python
test-python:
	$(MAKE) -C tests unit-tests

.PHONY: $(TESTDIRS)
$(TESTDIRS): test-%:
	$(MAKE) -C $* test

# Only install Go services and systemd unit files by default.
.PHONY: install
install: $(patsubst %,install-%,$(GODIRS) systemd)

.PHONY: install-java
install-java: $(JAVADIRS:%=install-%)

.PHONY: $(INSTALLDIRS)
$(INSTALLDIRS): install-%:
	$(MAKE) -C $* install

.PHONY: clean
clean: $(CLEANDIRS)
	rm -rf build dist

.PHONY: clean-java
clean-java: $(JAVADIRS:%=clean-%)

.PHONY: clean-go
clean-go: $(GODIRS:%=clean-%)

.PHONY: $(CLEANDIRS)
$(CLEANDIRS): clean-%:
	$(MAKE) -C $* clean

.PHONY: gopath
gopath:
	@echo $(GOPATH)

.PHONY: external
external:
	git submodule update --init

.PHONY: update-external
update-external:
	git submodule update --init --remote

.PHONY: version
version:
	python3 common/tools/update_project_version.py

# We cannot mark this target as phony without listing all possible targets.
%-dev:
	$(MAKE) $* DEVELOPMENT=1
