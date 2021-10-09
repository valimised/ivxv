rootdir := $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST))))../../)/
G := $(rootdir)/common/external/gradle-6.4/bin/gradle

.PHONY: all
all:
	$(G) build installDist $(if $(DEVELOPMENT),-P development=1) $(GFLAGS)

.PHONY: all-dev
all-dev:
	$(MAKE) DEVELOPMENT=1

.PHONY: test
test:
	$(G) test

.PHONY: install
install: GFLAGS := -x test
install: all
ifndef DESTDIR
	$(error $$DESTDIR must be set to install Java applications)
endif
	if [ -d build/distributions ]; then \
		mkdir -p $(DESTDIR)/ && \
		cp build/distributions/*.zip $(DESTDIR)/; \
	fi

.PHONY: sync
sync:
	SYNC_EXTERNAL=1 $(G) syncRemoteRepositories

.PHONY: clean
clean:
	$(G) clean
