help:
	@echo "usage: make all          Build service"
	@echo "usage: make clean          Clean build"

all:
	goreleaser release --skip-publish --rm-dist --snapshot

clean:
	rm -r dist