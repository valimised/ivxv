DOCS :=
DOCS := $(DOCS) protocols
DOCS := $(DOCS) architecture

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
	mkdir -p $(DESTDIR)/pdf/$*
	cp -ru $*/_build/latex/*.pdf $(DESTDIR)/pdf/$*

.PHONY: clean
clean: $(CLEANDOCS)
	rm -rf $(DESTDIR)

.PHONY: $(CLEANDOCS)
$(CLEANDOCS): clean-%:
	$(MAKE) -C $* clean
