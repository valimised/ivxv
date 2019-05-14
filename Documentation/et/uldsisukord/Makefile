mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

DEPENDENCY := nimekiri.rst

include ../../Makefile.common

.PHONY: ../../Makefile.common

nimekiri.rst:
	../../documents.py > $@

clean::
	rm -f nimekiri.rst
