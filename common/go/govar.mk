# govar.mk: Sets the GO environment variable to the location of the go binary.
# Intended to be included from MAkefiles that invoke go.

# Check the version of go on PATH. If newer than 1.7, then use that (useful for
# newer Go versions or non-Ubuntu platforms), otherwise explicitly specify the
# location of Go 1.7 on Ubuntu (our target OS). This is needed because even if
# you install golang-1.7-go on Ubuntu 16.04, /usr/bin/go will still point to Go
# 1.6 unless golang-go is also upgraded, which we do not require.
newer := $(shell echo "$$(go version | cut -d' ' -f3)\ngo1.7" | sort -V | tail -n1)
ifeq ($(newer),go1.7)
	export GO := /usr/lib/go-1.7/bin/go
else
	export GO := go
endif
