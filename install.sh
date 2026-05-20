#!/bin/bash
set -euo pipefail

# mongoman — Linux/macOS Installer
# Downloads or builds and installs mongoman to /usr/local/bin

APP="mongoman"
INSTALL_DIR="/usr/local/bin"

# Color helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info()  { echo -e "${GREEN}✅${NC} $1"; }
warn()  { echo -e "${YELLOW}⚠️${NC} $1"; }
error() { echo -e "${RED}❌${NC} $1"; }

# ── Pre-flight checks ─────────────────────────────────────────────────────────

# Check Go is installed
if ! command -v go &> /dev/null; then
    error "Go is not installed. Please install Go 1.22+ from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+' || echo "0.0")
GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)

if [ "$GO_MAJOR" -lt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 22 ]; }; then
    error "Go 1.22+ required (found $GO_VERSION)"
    exit 1
fi

info "Go $GO_VERSION detected"

# ── Build ─────────────────────────────────────────────────────────────────────

echo ""
echo "📦 Building $APP..."

BUILD_DIR=$(mktemp -d)
trap "rm -rf $BUILD_DIR" EXIT

# Determine Go architecture for build output
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
BINARY="$BUILD_DIR/$APP"

echo "   Target: $GOOS/$GOARCH"
echo ""

CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BINARY" .

if [ ! -f "$BINARY" ]; then
    error "Build failed"
    exit 1
fi

info "Build successful"

# ── Install ───────────────────────────────────────────────────────────────────

echo ""
echo "📋 Installing $APP to $INSTALL_DIR/$APP..."

if [ ! -d "$INSTALL_DIR" ]; then
    warn "$INSTALL_DIR does not exist, creating..."
    sudo mkdir -p "$INSTALL_DIR"
fi

sudo cp "$BINARY" "$INSTALL_DIR/$APP"
sudo chmod +x "$INSTALL_DIR/$APP"

info "Installed to $INSTALL_DIR/$APP"

# ── Verify ────────────────────────────────────────────────────────────────────

echo ""
if command -v $APP &> /dev/null; then
    info "$APP is now available. Run '$APP help' to get started."
    $APP help
else
    warn "$INSTALL_DIR/$APP is installed but may not be in your PATH."
    echo "   Add it: export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo ""
echo "🎉 Installation complete!"
