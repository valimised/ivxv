GO := /usr/lib/go-1.23/bin/go

# Only use go version >= 1.23
ifeq ($(shell which $(GO)),)
	fallback := $(shell which go)
	ifneq ($(fallback),)
		version := $(shell $(fallback) version | cut -d' ' -f3)
		newer := $(shell echo "go1.23\n$(version)" | sort --version-sort \
		| tail --lines=1)
		ifeq ($(version),$(newer))
			GO := $(fallback)
		endif
	endif
endif

export GO
