BINARY_NAME = buzz
LEGACY_BINARY_NAME = taproom
CLI_DIR = apps/cli
GO_MODULES = $(CLI_DIR) apps/desktop modules/brew modules/gh modules/protocol
GOBIN = $(HOME)/.local/bin
SOURCES = $(shell find $(GO_MODULES) -name '*.go')

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

install: build
	mkdir -p $(GOBIN)
	install -m 0755 $(BINARY_NAME) $(GOBIN)/$(BINARY_NAME)
	install -m 0755 $(LEGACY_BINARY_NAME) $(GOBIN)/$(LEGACY_BINARY_NAME)

test:
	@for module in $(GO_MODULES); do \
		echo "Testing $$module"; \
		profile="$$(mktemp "$${TMPDIR:-/tmp}/buzz-coverage.XXXXXX")" || exit $$?; \
		if ! go -C $$module test -count=1 -coverprofile=$$profile ./...; then rm -f $$profile; exit 1; fi; \
		coverage="$$(go tool cover -func=$$profile | awk '/^total:/ {print $$NF}')"; \
		if [ "$$coverage" != "100.0%" ]; then echo "Coverage gate failed for $$module: $${coverage:-no coverage}"; awk '$$NF == 0 {print "Uncovered Go source:", $$1}' $$profile; rm -f $$profile; exit 1; fi; \
		rm -f $$profile; \
		echo "Coverage gate passed for $$module: $$coverage"; \
	done

fmt:
	gofmt -w $(SOURCES)

fmt-check:
	@unformatted="$$(gofmt -l $(SOURCES))"; \
	if [ -n "$$unformatted" ]; then \
		echo "Go files need formatting:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	@for module in $(GO_MODULES); do \
		echo "Vetting $$module"; \
		go -C $$module vet ./... || exit $$?; \
	done

.PHONY: all build run clean install fmt fmt-check vet test
