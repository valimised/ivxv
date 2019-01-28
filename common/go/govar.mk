# govar.mk: Sets the GO environment variable to the location of the go binary.
# Intended to be included from Makefiles that invoke go.

fallback := /usr/lib/go-1.9/bin/go
GO := $(shell which go || echo $(fallback))

ifneq ($(GO),$(fallback))
	# Check the version of $(GO). If 1.9 or newer except for Go 1.10, then
	# use that (useful for newer Go versions or non-Ubuntu platforms),
	# otherwise use the fallback location. This is needed because even if
	# you install golang-1.9-go on Ubuntu, it will not update /usr/bin/go
	# (which is handled by golang-go).
	#
	# Explicitly exclude Go 1.10, because it has an issue that breaks X.509
	# with Estonian ID-cards and the fix done in Go 1.11 will no be
	# backported (see https://github.com/golang/go/issues/24151).
	minimal := go1.9
	exclude := go1.10
	gover := $(shell $(GO) version | cut -d' ' -f3)
	ifneq ($(gover),$(minimal))
		newer := $(shell echo "$(gover)\n$(minimal)" | sort -V | tail -n1)

		# If minimal is equal to newer or newer contains exclude then
		# use fallback location.
		ifneq (,$(or $(filter $(minimal),$(newer)),$(findstring $(exclude),$(newer))))
			GO := $(fallback)
		endif
	endif
endif

export GO # Export as environment variable for use in subshells.
