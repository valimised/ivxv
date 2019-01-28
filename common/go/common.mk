# common.mk: Common recipes for Go components.
#
# This Makefile is designed so that it can be both included and called,
# depending on if a recipe needs to be overridden or not.
#
# To use the recipes here unmodified, only an include is required. For example:
#
#     include ../common/go/gopath.mk
#     include ../common/go/common.mk
#
# However, if you wish to override some target, you can do so and set this file
# to be called from Go Makefiles as the fallback target. Example Makefile:
#
#     include ../common/go/gopath.mk
#
#     .DEFAULT_GOAL := all
#
#     .PHONY: clean
#     clean:
#     	# Custom clean recipe.
#
#     %: force
#     	$(MAKE) -f ../common/go/common.mk $@
#     force ../common/go/gopath.mk Makefile: ;
#
# Some notes about this example:
#
# The .DEFAULT_GOAL needs to be explicitly set, otherwise the Makefile defaults
# to cleaning.
#
# common.mk depends on the current working directory belonging to the calling
# Makefile, so -f (not -C) should be used when invoking it.
#
# The "force" target guarantees that the recipe will be run even if the target
# exists, without having to list all the targets in this Makefile as phony. It
# has an empty recipe to prevent recursion. See
# https://www.gnu.org/software/make/manual/html_node/Overriding-Makefiles.html
#
# The "../common/go/gopath.mk" and "Makefile" targets have empty recipes so
# that when make tries to automatically update them, they are not matched to
# the match-anything target.
#
# A target which is a prerequisite for another one cannot be overridden without
# also overriding the dependent target: it will use the version of the
# prerequisite which is in the same Makefile as the target.

COMMONDIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ROOTDIR   := $(abspath $(COMMONDIR)../../)/

include $(COMMONDIR)govar.mk

.PHONY: all
all: generate
	$(GO) install $(if $(DEVELOPMENT),-tags development )./...

.PHONY: lint
lint: all
	if which gometalinter > /dev/null; then \
		env PATH="$(dir $(GO)):$$PATH" TABWIDTH=8 \
			gometalinter --enable-all \
			--config $(COMMONDIR)gometalinter.json ./...; \
	fi

.PHONY: test
test: lint
	$(GO) test -v $(GOTESTFLAGS) ./...

.PHONY: install
install: all
	if [ -d bin ]; then mkdir -p $(DESTDIR)/usr && cp -r bin $(DESTDIR)/usr; fi
	$(if $(EXTRADATA),mkdir -p $(DESTDIR)/usr/share/ivxv && cp -r $(EXTRADATA) $(DESTDIR)/usr/share/ivxv)

.PHONY: generate
generate:
	$(GO) generate ./...
	$(ROOTDIR)common/tools/go/src/ivxv.ee/cmd/import ./...
# NOGEN is a special-case variable set only by common/tools/go/Makefile. It
# indicates that "gen" should not be built as this would result in a make
# recursion loop.
ifndef NOGEN
	$(MAKE) -C $(ROOTDIR) common/tools/go
	$(ROOTDIR)common/tools/go/bin/gen ./...
endif

.PHONY: clean
clean:
	find . -not -path \*/testdata/\* \( \
		-name gen_types.go -o \
		-name gen_types_test.go -o \
		-name gen_import.go -o \
		-name gen_import_dev.go \
		\) -delete
	$(GO) clean -i ./...
	rm -rf bin/ pkg/
