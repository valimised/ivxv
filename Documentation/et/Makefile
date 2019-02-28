DOCS :=
DOCS := $(DOCS) arhitektuur
DOCS := $(DOCS) haldusteenus
DOCS := $(DOCS) ivxvapi
DOCS := $(DOCS) kasutusmall
DOCS := $(DOCS) protokollid
DOCS := $(DOCS) seadistuste-koostejuhend
DOCS := $(DOCS) systeemihalduri-juhend
DOCS := $(DOCS) uldsisukord
DOCS := $(DOCS) votmerakendus

# following docs are excluded from the release
#DOCS := $(DOCS) entroopia
#DOCS := $(DOCS) eriomadused
#DOCS := $(DOCS) formal
#DOCS := $(DOCS) ivxv-pyapi
#DOCS := $(DOCS) regteenus
#DOCS := $(DOCS) testimisplaan
DOCS := $(DOCS) $(if $(DEVELOPMENT),tarne)
DESTDIR := _build

HTMLDOCS  := $(DOCS:%=html-%)
PDFDOCS   := $(DOCS:%=pdf-%)
CLEANDOCS := $(DOCS:%=clean-%)

.PHONY: all
all: html pdf

.PHONY: html
html: $(HTMLDOCS)

.PHONY: $(HTMLDOCS)
$(HTMLDOCS): html-%:
	$(MAKE) -C $* html
	mkdir -p $(DESTDIR)/html/$*
	cp -ru $*/_build/html/** $(DESTDIR)/html/$*

.PHONY: pdf
pdf: $(PDFDOCS)

.PHONY: $(PDFDOCS)
$(PDFDOCS): pdf-%:
	$(MAKE) -C $* latexpdf
	mkdir -p $(DESTDIR)/pdf
	cp -ru $*/_build/latex/*.pdf $(DESTDIR)/pdf/

.PHONY: clean
clean: $(CLEANDOCS)
	rm -rf $(DESTDIR)

.PHONY: $(CLEANDOCS)
$(CLEANDOCS): clean-%:
	$(MAKE) -C $* clean
