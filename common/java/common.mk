rootdir := $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST))))../../)/
G := $(rootdir)/common/external/gradle-3.0/bin/gradle

.PHONY: all
all:
	$(G) build installDist $(if $(DEVELOPMENT),-P development=1)

.PHONY: all-dev
all-dev:
	$(MAKE) DEVELOPMENT=1

.PHONY: test
test:
	$(G) test

.PHONY: clean
clean:
	$(G) clean
