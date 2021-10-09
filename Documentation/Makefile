# This Makefile supports two different invocation "modes".
#
# The default mode, invoked via "all", "pdf", or "html", builds all of the
# documentation with a few exceptions, putting PDF files into $(DESTDIR)/pdf
# (_build/pdf by default) and HTML files into $(DESTDIR)/html (_build/html by
# default, "all" builds html only in development mode).
#
# Release mode, invoked via "release", builds only the release documents and
# puts the PDF files, examples, and PDF diffs into a release hierarchy created
# at $(DESTDIR) (_build/release by default).

# Can be overridden from the command line.
DESTDIR = _build

# Build all Estonian and English documentation with a few exceptions.
DOCS := $(wildcard et/* en/*)
DOCS := $(filter-out et/example-config,$(DOCS)) # Does not contain reST.
DOCS := $(filter-out et/formal,$(DOCS))         # Does not contain reST.
DOCS := $(filter-out et/ivxv-pyapi,$(DOCS))     # Too many Python dependencies.
DOCS := $(filter-out et/entroopia,$(DOCS))      # Does not have a Makefile?

PDFDOCS   := $(DOCS:%=pdf-%)
HTMLDOCS  := $(DOCS:%=html-%)

.PHONY: all
all: pdf $(if $(DEVELOPMENT),html)

.PHONY: pdf
pdf: $(PDFDOCS)

.PHONY: $(PDFDOCS)
$(PDFDOCS): pdf-%:
	mkdir -p $(DESTDIR)/pdf
	$(MAKE) -C $* install-pdf DESTDIR=$(abspath $(DESTDIR)/pdf)

.PHONY: html
html: $(HTMLDOCS)

.PHONY: $(HTMLDOCS)
$(HTMLDOCS): html-%:
	mkdir -p $(DESTDIR)/html/$(notdir $*) # Drop language prefix.
	$(MAKE) -C $* install-html DESTDIR=$(abspath $(DESTDIR)/html/$(notdir $*))


# Rules for generating release documentation.

RELEASE1 = "$(abspath $(DESTDIR))/1. Ülddokumendid"
DOCS1 := et/uldsisukord

RELEASE2 = "$(abspath $(DESTDIR))/2. Spetsifikatsioonid"
DOCS2 := \
	et/arhitektuur \
	et/kasutusmall \
	et/protokollid \
	et/votmerakendus \
	en/architecture \
	en/protocols

RELEASE3 = "$(abspath $(DESTDIR))/3. Juhendid"
DOCS3 := \
	et/haldusteenus \
	et/ivxvapi \
	et/kogumisteenuse_haldusjuhend \
	et/seadistuste_koostejuhend \
	et/audiitor

RELEASE4 = "$(abspath $(DESTDIR))/4. Näited"
DOCS4 := \
	et/example-config \
	et/seadistuste_koostejuhend/config-examples/android-ios-config.json

RELEASE5 = "$(abspath $(DESTDIR))/5. Muudatused"

RELEASEDOCS := $(patsubst %,release-%,$(DOCS1) $(DOCS2) $(DOCS3) $(DOCS4))

# If releasing locally (i.e., no explicit DESTDIR is given), then use
# _build/release. RELEASE[1-5] must be recursively expanded for this to work.
$(RELEASEDOCS): DESTDIR := $(DESTDIR)/release

.PHONY: release
release: $(RELEASEDOCS)

.PHONY: $(RELEASEDOCS)
$(DOCS1:%=release-%): release-%:
	mkdir -p $(RELEASE1)
	$(MAKE) -C $* install-pdf DESTDIR=$(RELEASE1)
	mkdir -p $(RELEASE5)
	$(MAKE) -C $* install-diff DESTDIR=$(RELEASE5)

$(DOCS2:%=release-%): release-%:
	mkdir -p $(RELEASE2)
	$(MAKE) -C $* install-pdf DESTDIR=$(RELEASE2)
	mkdir -p $(RELEASE5)
	$(MAKE) -C $* install-diff DESTDIR=$(RELEASE5)

$(DOCS3:%=release-%): release-%:
	mkdir -p $(RELEASE3)
	$(MAKE) -C $* install-pdf DESTDIR=$(RELEASE3)
	mkdir -p $(RELEASE5)
	$(MAKE) -C $* install-diff DESTDIR=$(RELEASE5)

$(DOCS4:%=release-%): release-%:
	mkdir -p $(RELEASE4)
	cp --recursive --update $* $(RELEASE4)

# Cleanup.

# DOCS already contains DOCS[1-3].
CLEANDOCS := $(DOCS:%=clean-%)

.PHONY: clean
clean: $(CLEANDOCS)
	-git worktree remove --force _master
	rm -rf $(DESTDIR) _master
	py3clean .

.PHONY: $(CLEANDOCS)
$(CLEANDOCS): clean-%:
	$(MAKE) -C $* clean
