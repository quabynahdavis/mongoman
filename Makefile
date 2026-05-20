# mongoman — Cross-platform build system
#
# Builds a single static binary for each target platform.
# Output: build/<os>/<arch>/mongoman[.exe]

APP      := mongoman
MODULE   := github.com/davisdeveloper/mongoman
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE     := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -ldflags="-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE) -s -w"
OUTPUT   := build

# Default: build for the current platform
.PHONY: all
all: build

# ── Build current platform ────────────────────────────────────────────────────
.PHONY: build
build:
	@mkdir -p $(OUTPUT)
	go build $(LDFLAGS) -o $(OUTPUT)/$(APP) .

# ── Cross-compile for all supported platforms ─────────────────────────────────
PLATFORMS := linux/amd64 linux/arm64 linux/386 darwin/amd64 darwin/arm64 windows/amd64 windows/386 freebsd/amd64 openbsd/amd64 netbsd/amd64

.PHONY: cross
cross:
	@mkdir -p $(OUTPUT)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		dirname="$(OUTPUT)/$$os/$$arch"; \
		mkdir -p $$dirname; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(LDFLAGS) -o "$$dirname/$(APP)$$ext" .; \
	done
	@echo "✅ Cross-compilation complete. Output in $(OUTPUT)/"

# ── Install locally ───────────────────────────────────────────────────────────
.PHONY: install
install: build
	@echo "Installing $(APP) to /usr/local/bin/ (may require sudo)..."
	sudo cp $(OUTPUT)/$(APP) /usr/local/bin/$(APP)
	sudo chmod +x /usr/local/bin/$(APP)
	@echo "✅ Installed to /usr/local/bin/$(APP)"

# ── Clean ─────────────────────────────────────────────────────────────────────
.PHONY: clean
clean:
	rm -rf $(OUTPUT)
	@echo "Cleaned build output."

# ── Tests ─────────────────────────────────────────────────────────────────────
.PHONY: test
test:
	go test ./... -v

# ── Lint ──────────────────────────────────────────────────────────────────────
.PHONY: lint
lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# ── Release archives ──────────────────────────────────────────────────────────
.PHONY: release
release: cross
	@for dir in $(OUTPUT)/*/*/; do \
		os=$$(echo $$dir | cut -d'/' -f2); \
		arch=$$(echo $$dir | cut -d'/' -f3); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		archive="$(OUTPUT)/$(APP)_$${os}_$${arch}.tar.gz"; \
		echo "Packaging $$os/$$arch -> $$archive"; \
		cp README.md "$$dir/"; \
		cp LICENSE "$$dir/" 2>/dev/null || true; \
		cd "$$dir" && tar -czf "../../../$$archive" "$(APP)$$ext" README.md LICENSE 2>/dev/null; \
		cd ../../..; \
	done
	@echo "✅ Release archives created in $(OUTPUT)/"

# ── Help ──────────────────────────────────────────────────────────────────────
.PHONY: help
help:
	@echo "Targets:"
	@echo "  build    - Build for current platform"
	@echo "  cross    - Cross-compile for all supported platforms"
	@echo "  install  - Install to /usr/local/bin"
	@echo "  clean    - Remove build output"
	@echo "  test     - Run tests"
	@echo "  lint     - Run golangci-lint"
	@echo "  release  - Create release archives for all platforms"
	@echo "  help     - Show this message"
