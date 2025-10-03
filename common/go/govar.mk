GO := /usr/lib/go-1.23/bin/go
GOPATHLOCAL := $(ROOTDIR)common/external/go

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

# go downloads dependencies into GOMODCACHE
export GOMODCACHE=$(GOPATHLOCAL)

# go searches for dependencies declared in go.mod in:
# a) the Internet if GOPROXY is unset
# b) local directory if GOPROXY is set
ifndef ONLINE
export GOPROXY=file://$(GOPATHLOCAL)/cache/download
export GOSUMDB=off
endif
