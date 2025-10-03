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
DOCS := $(wildcard internal/* public/* et/*)
DOCS := $(filter-out et/example-config,$(DOCS)) # Does not contain reST.
DOCS := $(filter-out internal/formal,$(DOCS))         # Does not contain reST.
DOCS := $(filter-out internal/ivxv-pyapi,$(DOCS))     # Too many Python dependencies.
DOCS := $(filter-out internal/entroopia,$(DOCS))      # Does not have a Makefile?
DOCS := $(filter-out internal/stresstesting,$(DOCS))  # Does not have a Makefile?

PDFDOCS   := $(DOCS:%=pdf-%)
DIFFDOCS   := $(DOCS:%=diff-%)
HTMLDOCS  := $(DOCS:%=html-%)

.PHONY: all
all: release

.PHONY: pdf
pdf: $(PDFDOCS)
	$(MAKE) -C public/arhitektuur install-en-pdf DESTDIR=$(abspath $(DESTDIR)/pdf)
	$(MAKE) -C public/liidesed install-en-pdf DESTDIR=$(abspath $(DESTDIR)/pdf)
	$(MAKE) -C public/protokollid install-en-pdf DESTDIR=$(abspath $(DESTDIR)/pdf)

.PHONY: $(PDFDOCS)
$(PDFDOCS): pdf-%:
	mkdir -p $(DESTDIR)/pdf
	$(MAKE) -C $* install-pdf DESTDIR=$(abspath $(DESTDIR)/pdf)


.PHONY: diff
diff: $(DIFFDOCS)

.PHONY: $(DIFFDOCS)
$(DIFFDOCS): diff-%:
	mkdir -p $(DESTDIR)/diff
	$(MAKE) -C $* install-diff DESTDIR=$(abspath $(DESTDIR)/diff)


.PHONY: html
html: $(HTMLDOCS)

.PHONY: $(HTMLDOCS)
$(HTMLDOCS): html-%:
	mkdir -p $(DESTDIR)/html/$(notdir $*) # Drop language prefix.
	$(MAKE) -C $* install-html DESTDIR=$(abspath $(DESTDIR)/html/$(notdir $*))


# Rules for generating release documentation.

RELEASE1 = "$(abspath $(DESTDIR))/1. Ülddokumendid"
DOCS1ET := public/uldsisukord

RELEASE2 = "$(abspath $(DESTDIR))/2. Spetsifikatsioonid"
DOCS2ET := \
	public/arhitektuur \
	public/liidesed \
	et/kasutusmall \
	public/protokollid \
	et/xteeteenus \
	et/votmerakendus
DOCS2EN := \
	public/arhitektuur \
	public/liidesed \
	public/protokollid \
	en/backendlogs

RELEASE3 = "$(abspath $(DESTDIR))/3. Juhendid"
DOCS3ET := \
	et/haldusteenus \
	et/kogumisteenuse_haldusjuhend \
	et/seadistuste_koostejuhend \
	et/audiitor

RELEASE4 = "$(abspath $(DESTDIR))/4. Näited"
DOCS4ET := \
	et/example-config \
	et/seadistuste_koostejuhend/config-examples/android-ios-config.json

RELEASE5 = "$(abspath $(DESTDIR))/5. Muudatused"

RELEASEDOCSET := $(patsubst %,release-et-%,$(DOCS1ET) $(DOCS2ET) $(DOCS3ET) $(DOCS4ET))
RELEASEDOCSEN := $(patsubst %,release-en-%,$(DOCS1EN) $(DOCS2EN))

# If releasing locally (i.e., no explicit DESTDIR is given), then use
# _build/release. RELEASE[1-5] must be recursively expanded for this to work.
$(RELEASEDOCSET): DESTDIR := $(DESTDIR)/release
$(RELEASEDOCSEN): DESTDIR := $(DESTDIR)/release

.PHONY: release
release: $(RELEASEDOCSET) $(RELEASEDOCSEN)

.PHONY: $(RELEASEDOCSET)
$(DOCS1ET:%=release-et-%): release-et-%:
	mkdir -p $(RELEASE1)
	$(MAKE) -C $* install-pdf DESTDIR=$(RELEASE1)
	mkdir -p $(RELEASE5)
	$(MAKE) -C $* install-diff DESTDIR=$(RELEASE5)

$(DOCS2ET:%=release-et-%): release-et-%:
	mkdir -p $(RELEASE2)
	$(MAKE) -C $* install-pdf DESTDIR=$(RELEASE2)
	mkdir -p $(RELEASE5)
	$(MAKE) -C $* install-diff DESTDIR=$(RELEASE5)

$(DOCS3ET:%=release-et-%): release-et-%:
	mkdir -p $(RELEASE3)
	$(MAKE) -C $* install-pdf DESTDIR=$(RELEASE3)
	mkdir -p $(RELEASE5)
	$(MAKE) -C $* install-diff DESTDIR=$(RELEASE5)

$(DOCS4ET:%=release-et-%): release-et-%:
	mkdir -p $(RELEASE4)
	cp --recursive --update $* $(RELEASE4)


.PHONY: $(RELEASEDOCSEN)
$(DOCS2EN:%=release-en-%): release-en-%:
	mkdir -p $(RELEASE2)
	$(MAKE) -C $* install-en-pdf DESTDIR=$(RELEASE2)


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
