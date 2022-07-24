# Set DEB_BUILD_OPTIONS=nocheck in development branch to disable unit tests during packaging
DEB_BUILD_OPTIONS:=$(shell test "$$(git rev-parse --abbrev-ref HEAD)" = develop && echo "nocheck" || true )
DOCKER_OPTS=--tty --rm=true $(if $(DEVELOPMENT),--env DEVELOPMENT=$(DEVELOPMENT))

IVXV_VERSION:=$(shell head -1 ../debian/changelog | sed -r -e 's/.+\((.+)\).+/\1/')
IVXV_RELEASE_DIR:=ivxv-$(IVXV_VERSION)
APPS_RELEASE_DIR:=ivxv-apps-$(IVXV_VERSION)

# build and test IXVV release
.PHONY: release
release: check-release-status build-ivxv-packages
	# initialize output directory
	rsync -a --del template/ ivxv-release-build/

	# move IVXV packages to release
	mv --verbose ivxv-debs-build ivxv-release-build/debs


	# build documentation
	$(MAKE) build-docs
	mv ivxv-docs-build ivxv-release-build/docs

	# output release data
	@test ../.git && \
		echo "IVXV repository branch: $(shell git rev-parse --abbrev-ref HEAD)\n" \
		> ivxv-release-build/build.txt
	cd ivxv-release-build \
		&& find -type f ! -name sha256sum -exec sha256sum {} \; \
		| tee sha256sum | sha256sum --check --quiet

	rm --recursive --force $(IVXV_RELEASE_DIR)
	mv ivxv-release-build $(IVXV_RELEASE_DIR)

	@echo "================================"
	@echo "IVXV package version: $(IVXV_VERSION)"
	@echo "Release directory: release/$(IVXV_RELEASE_DIR)/"
	@echo "================================"

.PHONY: release-apps
release-apps: build-apps
	rm --force --recursive ivxv-apps-$(IVXV_VERSION)
	mv ivxv-apps-build ivxv-apps-$(IVXV_VERSION)
	@echo "================================"
	@echo "IVXV apps version: $(IVXV_VERSION)"
	@echo "Release directory: release/ivxv-apps-$(IVXV_VERSION)/"
	@echo "================================"

.PHONY: release-src
release-src:
	git -C .. archive --format zip --output release/ivxv-src-build.zip.tmp $(IVXV_VERSION)
	zip -d ivxv-src-build.zip.tmp .gitmodules release/Jenkinsfile
	zip ivxv-src-build.zip.tmp ../common/external/README.rst
	mv --verbose ivxv-src-build.zip.tmp ivxv_$(IVXV_VERSION).source.zip

# build docker image for packaging
docker-image:
	docker build -t ivxv-build:latest $(IVXV_DOCKER_BUILD_ARGS) docker

# check release status of changelog file (if not in development branch)
check-release-status:
	@if head -1 ../debian/changelog | egrep UNRELEASED; then ( \
		if test "$$(git rev-parse --abbrev-ref HEAD)" != develop ; then ( \
			echo '^^^^^^^^^^^^^^^^^^^^^^^^^^^^'; \
			echo 'Debian distribution in IVXV changelog is UNRELEASED!!!'; \
			echo; \
			echo 'Before actual release change it to "bionic"'; \
			echo 'Execute "dch --release" in project root directory'; \
			echo; \
			echo 'Sleeping 10 seconds...'; \
			sleep 10; \
		); fi \
	); fi

# build IVXV packages
.PHONY: build-ivxv-packages
build-ivxv-packages: docker-image
	rm --recursive --force ivxv-debs-build
	mkdir ivxv-debs-build
	docker run $(DOCKER_OPTS) \
		--volume $(CURDIR)/ivxv-debs-build:/home/jenkins/src \
		--volume $(CURDIR)/..:/home/jenkins/src/ivxv \
		--user $(shell id --user):$(shell id --group) \
		--env DEB_BUILD_OPTIONS=$(DEB_BUILD_OPTIONS) \
		--workdir /home/jenkins/src/ivxv \
		-- \
		ivxv-build:latest \
		dpkg-buildpackage -b -uc
	rmdir ivxv-debs-build/ivxv

# build documentation
.PHONY: build-docs
build-docs: docker-image
	docker run $(DOCKER_OPTS) \
		--volume $(CURDIR)/..:/home/jenkins/src/ivxv \
		--user $(shell id --user):$(shell id --group) \
		--workdir /home/jenkins/src/ivxv/Documentation/common/schema \
		-- \
		ivxv-build:latest \
		make validate
	rm --recursive --force ivxv-docs-build
	docker run $(DOCKER_OPTS) \
		--volume $(CURDIR)/..:/home/jenkins/src/ivxv \
		--user $(shell id --user):$(shell id --group) \
		--workdir /home/jenkins/src/ivxv/Documentation \
		-- \
		ivxv-build:latest \
		make release DESTDIR=/home/jenkins/src/ivxv/release/ivxv-docs-build

# build applications
.PHONY: build-apps
build-apps: docker-image
	rm --recursive --force ivxv-apps-build

	# build applications
	docker run $(DOCKER_OPTS) \
		--volume $(CURDIR)/..:/home/jenkins/src/ivxv \
		--user $(shell id --user):$(shell id --group) \
		--workdir /home/jenkins/src/ivxv \
		-- \
		ivxv-build:latest \
		make install-java DESTDIR=/home/jenkins/src/ivxv/release/ivxv-apps-build

	# build test applications (key)
	docker run $(DOCKER_OPTS) \
		--volume $(CURDIR)/..:/home/jenkins/src/ivxv \
		--user $(shell id --user):$(shell id --group) \
		--workdir /home/jenkins/src/ivxv \
		-- \
		ivxv-build:latest \
		make install-key-dev DESTDIR=/home/jenkins/src/ivxv/release/ivxv-apps-build/testapps
	mv ivxv-apps-build/testapps/key-$(IVXV_VERSION).zip ivxv-apps-build/testapps/key-test-$(IVXV_VERSION).zip

# clean release directory and untag docker image
.PHONY: clean
clean:
	rm --recursive --force debs-* docs apps
	rm --recursive --force ivxv-* ivxv_* output
	docker rmi --force --no-prune ivxv-build:latest
