DOCS :=
DOCS := $(DOCS) architecture
DOCS := $(DOCS) protocols

DESTDIR := _build
PREVIVXVDOC := $(DESTDIR)/ivxv/Documentation/en

OLDPDFDOCS    := $(DOCS:%=oldpdf-%)
DIFFDOCS      := $(DOCS:%=diff-%)
HTMLDOCS      := $(DOCS:%=html-%)
PDFDOCS       := $(DOCS:%=pdf-%)
CLEANDOCS     := $(DOCS:%=clean-%)

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
	py3clean .
	rm -rf $(DESTDIR)

.PHONY: $(CLEANDOCS)
$(CLEANDOCS): clean-%:
	$(MAKE) -C $* clean


.PHONY: diff
diff: $(DIFFDOCS)
	find . -name \*-diff.tex -print0 | xargs -0 grep -v PREAMBLE | grep -E '(DIFdel|DIFadd)' | cut -d : -f 1 | uniq -c


$(PREVIVXVDOC):
	cd $(DESTDIR) && git clone git:/opt/git/ivxv


.PHONY: $(OLDPDFDOCS)
$(OLDPDFDOCS): oldpdf-%: $(PREVIVXVDOC)
	cd $(PREVIVXVDOC) && make pdf-$*


.PHONY: $(DIFFDOCS)
$(DIFFDOCS): diff-%: pdf-% oldpdf-%
	../differ.sh $*

