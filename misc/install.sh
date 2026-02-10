#!/bin/bash
set -e

# Usage: curl -sSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | bash
# Optional: GNO_DIR=/custom/path curl -sSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | bash
# Uninstall: curl -sSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | bash -s -- --uninstall
#
# This script is temporarily located in misc/ as we expect more official installation
# methods to emerge. It provides a convenient one-liner for installing gno, which is
# particularly useful when working with go.mod files containing replace directives
# that might conflict with direct `go install` commands.

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
log() {
    echo -e "${GREEN}[gno-install]${NC} $1"
}

error() {
    echo -e "${RED}[gno-install]${NC} $1" >&2
}

warn() {
    echo -e "${YELLOW}[gno-install]${NC} $1"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to determine gno source directory
get_gno_dir() {
    if [ -n "$GNO_DIR" ]; then
        echo "$GNO_DIR"
    elif [ -n "$HOME" ]; then
        echo "$HOME/.gno/src"
    else
        echo "/usr/local/share/gno"
    fi
}

# Function to check Go installation
check_go() {
    if ! command_exists go; then
        error "Go is not installed. Please install Go first:"
        echo "  https://golang.org/doc/install"
        exit 1
    fi

    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [ "$(echo "$GO_VERSION 1.18" | awk '{print ($1 < $2)}')" -eq 1 ]; then
        error "Go version 1.18 or higher is required. Current version: $GO_VERSION"
        exit 1
    fi
}

# Function to install gno
install_gno() {
    local GNO_DIR
    GNO_DIR=$(get_gno_dir)

    if ! command_exists git; then
      error "git is not installed. Please install git first."
      exit 1
    fi

    log "Installing gno source to $GNO_DIR"

    mkdir -p "$GNO_DIR"
    # Clone or update repository
    if [ -d "$GNO_DIR/.git" ]; then
        log "Updating existing gno repository..."
        cd "$GNO_DIR"
        git fetch --depth 1
        git reset --hard origin/master
    else
        log "Cloning gno repository..."
        git clone --depth 1 https://github.com/gnolang/gno.git "$GNO_DIR"
        cd "$GNO_DIR"
    fi

    # Build and install
    log "Building gno, gnokey, gnodev..."
    make install
    
    # Build and install gnobro
    log "Building gnobro..."
    make install.gnobro

    # Verify installation
    if ! command_exists gno; then
        error "Installation failed. gno command not found."
        log "Is $GOBIN set in your $PATH? See https://go.dev/doc/install/source#environment"
        exit 1
    fi
    
    if ! command_exists gnobro; then
        warn "gnobro installation failed. gnobro command not found."
        warn "You can install it manually later with: make install.gnobro"
    fi

    log "Installation successful! gno is now available."
    gno version
    
    if command_exists gnobro; then
        log "gnobro is also available."
    fi
}

# Function to uninstall gno
uninstall_gno() {
    local GNO_DIR
    local GOPATH
    GNO_DIR=$(get_gno_dir)
    GOPATH=$(go env GOPATH)

    log "Uninstalling gno binaries from $GOPATH/bin"
    rm -f "$GOPATH/bin/gno"
    rm -f "$GOPATH/bin/gnokey"
    rm -f "$GOPATH/bin/gnodev"
    rm -f "$GOPATH/bin/gnobro"

    # Remove source directory
    log "Removing gno source from $GNO_DIR"
    rm -rf "$GNO_DIR"

    log "Uninstallation complete."
}

# Main script
if [ "$1" = "--uninstall" ]; then
    uninstall_gno
    exit 0
fi

# Check Go installation
check_go

# Install gno
install_gno
