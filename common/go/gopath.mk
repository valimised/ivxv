# gopath.mk: Sets the GOPATH environment variable to the location of all Go
# components. Intended to be included from Makefiles that require a GOPATH.

rootdir := $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST))))../../)/
paths := common/external/gopath common/tools/go common/collector choices dds proxy storage verification voting

# Small Makefile trick to get a variable containing a space.
# http://stackoverflow.com/a/1542661
noop=
space = $(noop) $(noop)

export GOPATH := $(subst $(space),:,$(addprefix $(rootdir),$(paths)))
