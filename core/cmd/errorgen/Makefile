include ../../govar.mk

.PHONY: all
all: bin/gen

bin/gen: clean
	$(GO) generate ./...
	env GOBIN=$(or $(TIVI_GOBIN),$(abspath bin)) \
		$(GO) install $(if $(DEVELOPMENT),-tags development )./...

.PHONY: clean
clean:
	rm -rf cmd/gen/gentmpl.go
