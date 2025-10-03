G := $(ROOTDIR)common/external/gradle-8.11/bin/gradle
G_CACHE := $(ROOTDIR)common/external/java/
GFLAGS := -g=$(G_CACHE)

# Will download Java dependencies from Web into common/external/java/
# if you run from project's root `make [...] ONLINE=[...]`, otherwise
# local Java dependencies from common/external/java/ will be used
ifndef ONLINE
GFLAGS += --offline
endif

