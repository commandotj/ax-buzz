BINARY_NAME = buzz
LEGACY_BINARY_NAME = taproom
CLI_DIR = apps/cli
GOBIN = $(HOME)/.local/bin
TARGET_OS = darwin
TARGET_ARCH = arm64 amd64
VERSION = $(shell cat $(CLI_DIR)/.version)
ASSETS  = $(wildcard release/*)
SOURCES = $(shell find $(CLI_DIR) -name '*.go')
RELEASE_NOTE = .release-note.md

all: build

$(BINARY_NAME): $(SOURCES) $(CLI_DIR)/.version
	go -C $(CLI_DIR) build -o ../../$(BINARY_NAME) .

$(LEGACY_BINARY_NAME): $(BINARY_NAME)
	cp $(BINARY_NAME) $(LEGACY_BINARY_NAME)

build: $(BINARY_NAME) $(LEGACY_BINARY_NAME)

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME) $(LEGACY_BINARY_NAME)
	rm -rf release

install: build
	mkdir -p $(GOBIN)
	install -m 0755 $(BINARY_NAME) $(GOBIN)/$(BINARY_NAME)
	install -m 0755 $(LEGACY_BINARY_NAME) $(GOBIN)/$(LEGACY_BINARY_NAME)

test:
	go -C $(CLI_DIR) test ./...

fmt:
	gofmt -w $(shell find $(CLI_DIR) -name '*.go')

vet:
	go -C $(CLI_DIR) vet ./...

release:
	mkdir -p release
	for arch in $(TARGET_ARCH); do \
		echo "Building release/$(BINARY_NAME)-$(TARGET_OS)-$$arch"; \
		GOOS=$(TARGET_OS) GOARCH=$$arch go -C $(CLI_DIR) build -o ../../release/$(BINARY_NAME)-$(TARGET_OS)-$$arch .; \
		cp release/$(BINARY_NAME)-$(TARGET_OS)-$$arch release/$(LEGACY_BINARY_NAME)-$(TARGET_OS)-$$arch; \
		tar -C release -czf release/$(BINARY_NAME)-$(TARGET_OS)-$$arch.tar.gz $(BINARY_NAME)-$(TARGET_OS)-$$arch; \
		tar -C release -czf release/$(LEGACY_BINARY_NAME)-$(TARGET_OS)-$$arch.tar.gz $(LEGACY_BINARY_NAME)-$(TARGET_OS)-$$arch; \
		rm release/$(BINARY_NAME)-$(TARGET_OS)-$$arch release/$(LEGACY_BINARY_NAME)-$(TARGET_OS)-$$arch; \
	done
	(cd release && shasum -a 256 *.tar.gz > checksum.txt)

gh-release: release
	git tag $(VERSION)
	git push origin $(VERSION)
	$(EDITOR) $(RELEASE_NOTE)
	gh release create $(VERSION) $(ASSETS) \
		--title "$(VERSION)" \
		--notes-file $(RELEASE_NOTE)
	rm $(RELEASE_NOTE)

.PHONY: all build run clean install fmt vet test release
