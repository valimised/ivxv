SERVICES := choices mid proxy storage voting verification
OUTPUT   := $(SERVICES:%=ivxv-%@.service)

.PHONY: all
all: $(OUTPUT)

ivxv-%@.service: ivxv-$$service@.service.in
	sed 's/@service@/$*/' < '$^' > $@

.PHONY: install
install: all
	mkdir -p $(DESTDIR)/usr/lib/systemd/user
	cp $(OUTPUT) $(DESTDIR)/usr/lib/systemd/user

.PHONY: clean
clean:
	rm -f $(OUTPUT)
