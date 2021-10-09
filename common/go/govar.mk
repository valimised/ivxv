# govar.mk: Sets the GO environment variable to the location of the go binary.
# Intended to be included from Makefiles that invoke go.

GO := /usr/lib/go-1.14/bin/go

# Prefer to use the exact Go version, but if it does not exist then attempt to
# fallback to a go binary on PATH, given it is at least version 1.14.
ifeq ($(shell which $(GO)),)
	fallback := $(shell which go)
	ifneq ($(fallback),)
		version := $(shell $(fallback) version | cut -d' ' -f3)
		newer := $(shell echo "go1.14\n$(version)" | sort --version-sort | tail --lines=1)
		ifeq ($(version),$(newer))
			GO := $(fallback)
		endif
	endif
endif

export GO # Export as environment variable for use in subshells.
