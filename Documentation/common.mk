# common.mk: Common recipes for Documentation.
#
# Defines a catch-all recipe which routes unknown targets to Sphinx in "make
# mode". If $(DEPENDENCIES) is defined, then it will be set as a prerequisite
# of the catch-all and removed on clean. In addition, if the source directory
# contains a "model" subdirectory, then it invokes make in it to build and
# clean that too.
#
# Defines a "diff" recipe which generates a PDF that highlights differences
# between the current and latest released version of the document. If there are
# no differences, then no PDF is produced.
#
# Defines "install-pdf", "install-html", and "install-diff" recipes which build
# the specified type of documentation and install it to $(DESTDIR).

common.mk := $(lastword $(MAKEFILE_LIST))
Makefile  := $(lastword $(filter-out $(common.mk),$(MAKEFILE_LIST)))

# You can set these variables from the command line, and also
# from the environment for the first two.
SPHINXOPTS    ?= -W -c $(dir $(common.mk))
SPHINXBUILD   ?= sphinx-build
SOURCEDIR     = $(dir $(Makefile))
BUILDDIR      = $(SOURCEDIR)_build

# Set IVXV_DOCUMENT to the name of the source directory. This will be used as
# the key to look up configuration from documents.py.
export IVXV_DOCUMENT := $(notdir $(patsubst %/,%,$(abspath $(SOURCEDIR))))

# Special case the help target: set as default and skip prerequisites.
.DEFAULT_GOAL := help
.PHONY: help
help:
	@$(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

# Special case the clean target: skip prerequisites and perform extra steps.
.PHONY: clean
clean:
	#if [ -d "model" ]; then $(MAKE) -C model clean; fi
	@$(SPHINXBUILD) -M clean "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)
	rm -rf $(BUILDDIR) $(DEPENDENCIES)

# Do not regenerate the Makefiles.
$(common.mk) $(Makefile): ;

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: $(DEPENDENCIES)
	if [ -d "model" ]; then $(MAKE) -C model; fi
	$(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

# Variables and rules for diffing.
differ.sh := $(dir $(common.mk))differ.sh
master    := $(dir $(common.mk))_master/
gitdir    := $(patsubst $(shell git rev-parse --show-toplevel)%,%,$(abspath $(SOURCEDIR)))

.PHONY: diff
diff: latex master-latex
	# latexmk attempts to proceed even though errors may be present
	# pdflatex attempts to proceed with errors and is silent (batchmode)
	LATEXMKOPTS="-f -interaction=batchmode" $(differ.sh) $(BUILDDIR)/master $(BUILDDIR)

.PHONY: master-latex
master-latex: $(master)
	if [ -d $(master)$(gitdir) ]; then \
		$(MAKE) -C $(master)$(gitdir) BUILDDIR=$(abspath $(BUILDDIR))/master latex; \
	fi

$(master):
	git worktree add $@ 1.7.7

# Installation rules.
.PHONY: install-pdf
install-pdf: latexpdf
	cp --update $(filter-out %-diff.pdf,$(wildcard $(BUILDDIR)/latex/*.pdf)) "$(DESTDIR)"

.PHONY: install-html
install-html: html
	cp --recursive --update $(BUILDDIR)/html/** "$(DESTDIR)"

.PHONY: install-diff
install-diff: diff
	$(eval diff.pdf = $(wildcard $(BUILDDIR)/latex/*-diff.pdf))
	$(if $(diff.pdf),cp --update $(diff.pdf) "$(DESTDIR)")
