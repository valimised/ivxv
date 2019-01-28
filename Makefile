include common/go/gopath.mk

JAVADIRS := common/java key processor auditor
GODIRS   := common/tools/go common/collector proxy dds choices voting verification storage
DOCS     := Documentation/et

DIRS        := $(JAVADIRS) $(GODIRS) systemd
HTMLDOCS    := $(DOCS:%=html-%)
PDFDOCS     := $(DOCS:%=pdf-%)
TESTDIRS    := $(patsubst %,test-%,$(JAVADIRS) $(GODIRS))
INSTALLDIRS := $(patsubst %,install-%,$(GODIRS) systemd)
CLEANDIRS   := $(patsubst %,clean-%,$(DIRS) $(DOCS))

export ROOT_BUILD=true

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

.PHONY: $(TESTDIRS)
$(TESTDIRS): test-%:
	$(MAKE) -C $* test

.PHONY: doc
doc: $(DOCS)

.PHONY: $(DOCS)
.SECONDEXPANSION: $(DOCS)
$(DOCS): $(if $(DEVELOPMENT),html-$$@) pdf-$$@

.PHONY: html
html: $(HTMLDOCS)

.PHONY: $(HTMLDOCS)
$(HTMLDOCS): html-%:
	$(MAKE) -C $* html

.PHONY: pdf
pdf: $(PDFDOCS)

.PHONY: $(PDFDOCS)
$(PDFDOCS): pdf-%:
	$(MAKE) -C $* pdf

.PHONY: install
install: $(INSTALLDIRS)

.PHONY: $(INSTALLDIRS)
$(INSTALLDIRS): install-%:
	$(MAKE) -C $* install

.PHONY: clean
clean: $(CLEANDIRS)
	rm -rf build ivxv-*

.PHONY: clean-java
clean-java: $(JAVADIRS:%=clean-%)

.PHONY: clean-go
clean-go: $(GODIRS:%=clean-%)

.PHONY: clean-doc
clean-doc: $(DOCS:%=clean-%)

.PHONY: $(CLEANDIRS)
$(CLEANDIRS): clean-%:
	$(MAKE) -C $* clean

.PHONY: gopath
gopath:
	@echo $(GOPATH)

# We cannot mark this target as phony without listing all possible targets.
%-dev:
	$(MAKE) $* DEVELOPMENT=1

