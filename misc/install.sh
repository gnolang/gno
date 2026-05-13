#!/bin/bash
set -e

# Usage: curl -sSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | bash [-s -- [--gno] [--gnokey] [--gnodev] [--gnobro]]
# Optional: GNOROOT=/custom/path curl -sSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | bash
# Uninstall: curl -sSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | bash -s -- --uninstall
# If no flags are provided, all tools will be installed
#
# This script is temporarily located in misc/ as we expect more official installation
# methods to emerge. It provides a convenient one-liner for installing gno, which is
# particularly useful when working with go.mod files containing replace directives
# that might conflict with direct `go install` commands.

# Tool (un)installation flags
INSTALL_GNO=true
INSTALL_GNOKEY=true
INSTALL_GNODEV=true
INSTALL_GNOBRO=true
UNINSTALL=false

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
get_gno_root() {
    if [ -n "$GNOROOT" ]; then
        echo "$GNOROOT"
    elif [ -n "$HOME" ]; then
        echo "$HOME/.gno/src"
    else
        echo "/usr/local/share/gno"
    fi
}

# Function to determine where go install binaries
# When using `go install` (as done in the gnolang/gno Makefiles),
# go will install binaries to $GOBIN if set, otherwise to $GOPATH/bin.
# `go env GOBIN` does not return a default value, so we need to check both.
# A lot of issues exists on the Go GitHub about this, e.g.: https://github.com/golang/go/issues/34522
get_go_bin() {
    local gobin="$(go env GOBIN 2>/dev/null || echo '')"
    local gopath="$(go env GOPATH 2>/dev/null || echo '')"

    if [ -n "$gobin" ]; then
        echo "$gobin"
    elif [ -n "$gopath" ]; then
        echo "$gopath/bin"
    else
      error "Could not determine Go binary installation path. Please ensure Go is installed correctly."
      exit 1
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
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    if [ "$(echo "$go_version 1.18" | awk '{print ($1 < $2)}')" -eq 1 ]; then
        error "Go version 1.18 or higher is required. Current version: $go_version"
        exit 1
    fi
}

# Parse command line arguments
parse_args() {
    if [ $# -eq 0 ]; then
        return
    fi

    # Reset all install flags to false
    read -r INSTALL_GNO INSTALL_GNOKEY INSTALL_GNODEV INSTALL_GNOBRO <<< 'false false false false'

    for arg in "$@"; do
        case $arg in
            --gno)
                INSTALL_GNO=true
                ;;
            --gnokey)
                INSTALL_GNOKEY=true
                ;;
            --gnodev)
                INSTALL_GNODEV=true
                ;;
            --gnobro)
                INSTALL_GNOBRO=true
                ;;
            --uninstall)
                UNINSTALL=true
                ;;
            *)
                error "Unknown flag: $arg"
                error "Valid flags are: --gno, --gnokey, --gnodev, --gnobro, --uninstall"
                exit 1
                ;;
        esac
    done
}

# Function to install gno
install_gno() {
    local gnoroot=$(get_gno_root)

    if ! command_exists git; then
      error "git is not installed. Please install git first."
      exit 1
    fi

    log "Installing gno source to $gnoroot"

    mkdir -p "$gnoroot"
    # Clone or update repository
    if [ -d "$gnoroot/.git" ]; then
        log "Updating existing gno repository..."
        cd "$gnoroot"
        git fetch --depth 1
        git reset --hard origin/master
    else
        log "Cloning gno repository..."
        git clone --depth 1 https://github.com/gnolang/gno.git "$gnoroot"
        cd "$gnoroot"
    fi

    # Build and install Gno tools
    if [ "$INSTALL_GNO" = true ]; then
        log "Building gno..."
        make install.gno
    fi

    if [ "$INSTALL_GNOKEY" = true ]; then
        log "Building gnokey..."
        make install.gnokey
    fi

    if [ "$INSTALL_GNODEV" = true ]; then
        log "Building gnodev..."
        make install.gnodev
    fi

    if [ "$INSTALL_GNOBRO" = true ]; then
        log "Building gnobro..."
        make install.gnobro
    fi

    # Verify installation
    local failed_tools=()

    if [ "$INSTALL_GNO" = true ] && ! command_exists gno; then
        failed_tools+=("gno")
    fi

    if [ "$INSTALL_GNOKEY" = true ] && ! command_exists gnokey; then
        failed_tools+=("gnokey")
    fi

    if [ "$INSTALL_GNODEV" = true ] && ! command_exists gnodev; then
        failed_tools+=("gnodev")
    fi

    if [ "$INSTALL_GNOBRO" = true ] && ! command_exists gnobro; then
        failed_tools+=("gnobro")
    fi

    if [ ${#failed_tools[@]} -gt 0 ]; then
        error "Installation failed. The following tools were not found: ${failed_tools[*]}"
        log "Is \$GOBIN set in your \$PATH? See https://go.dev/doc/install/source#environment"
        exit 1
    fi

    log "Installation complete."
}

# Function to uninstall gno
uninstall_gno() {
    local gnoroot=$(get_gno_root)
    local gobin=$(get_go_bin)

    log "Uninstalling gno binaries from $gobin"
    rm -f "$gobin/gno"
    rm -f "$gobin/gnokey"
    rm -f "$gobin/gnodev"
    rm -f "$gobin/gnobro"

    # Remove source directory
    log "Removing gno source from $gnoroot"
    rm -rf "$gnoroot"

    log "Uninstallation complete."
}

# Parse arguments and validate flags
parse_args "$@"

# Check Go installation
check_go

# Install or uninstall Gno tools
if [ "$UNINSTALL" = true ]; then
    uninstall_gno
else
    install_gno
fi
