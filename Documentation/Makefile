DOCS1 :=
DOCS1 := $(DOCS1) et/uldsisukord

DOCS2 :=
DOCS2 := $(DOCS2) et/arhitektuur
DOCS2 := $(DOCS2) et/kasutusmall
DOCS2 := $(DOCS2) et/protokollid
DOCS2 := $(DOCS2) et/votmerakendus
DOCS2 := $(DOCS2) en/architecture
DOCS2 := $(DOCS2) en/protocols

DOCS3 := $(DOCS3) et/haldusteenus
DOCS3 := $(DOCS3) et/ivxvapi
DOCS3 := $(DOCS3) et/seadistuste-koostejuhend
DOCS3 := $(DOCS3) et/systeemihalduri-juhend

DESTDIR := _build
DESTDIR1 := "$(DESTDIR)/1. Ülddokumendid"
DESTDIR2 := "$(DESTDIR)/2. Spetsifikatsioonid"
DESTDIR3 := "$(DESTDIR)/3. Juhendid"

PDFDOCS1   := $(DOCS1:%=pdf1-%)
PDFDOCS2   := $(DOCS2:%=pdf2-%)
PDFDOCS3   := $(DOCS3:%=pdf3-%)

CLEANDOCS :=
CLEANDOCS := $(CLEANDOCS) $(DOCS1:%=clean-%)
CLEANDOCS := $(CLEANDOCS) $(DOCS2:%=clean-%)
CLEANDOCS := $(CLEANDOCS) $(DOCS3:%=clean-%)


.PHONY: all
all: pdf

.PHONY: pdf
pdf: $(PDFDOCS1) $(PDFDOCS2) $(PDFDOCS3)

.PHONY: $(PDFDOCS1)
$(PDFDOCS1): pdf1-%:
	$(MAKE) -C $* latexpdf
	mkdir -p $(DESTDIR1)
	cp -ru $*/_build/latex/*.pdf $(DESTDIR1)

.PHONY: $(PDFDOCS2)
$(PDFDOCS2): pdf2-%:
	$(MAKE) -C $* latexpdf
	mkdir -p $(DESTDIR2)
	cp -ru $*/_build/latex/*.pdf $(DESTDIR2)

.PHONY: $(PDFDOCS3)
$(PDFDOCS3): pdf3-%:
	$(MAKE) -C $* latexpdf
	mkdir -p $(DESTDIR3)
	cp -ru $*/_build/latex/*.pdf $(DESTDIR3)

.PHONY: clean
clean: $(CLEANDOCS)
	rm -rf $(DESTDIR)

.PHONY: $(CLEANDOCS)
$(CLEANDOCS): clean-%:
	$(MAKE) -C $* clean
