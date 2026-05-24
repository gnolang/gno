#!/bin/sh
# Gno binary uninstaller.
# Run with --help for usage.

set -eu

FULL_COMPONENTS="gno gnokey gnodev gnobro gnoweb gnoland"
INSTALL_DIR="${GNO_INSTALL_DIR:-${HOME}/.gno/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log() { printf '%b[gno-uninstall]%b %s\n' "$GREEN" "$NC" "$1"; }
die() { printf '%b[gno-uninstall] error:%b %s\n' "$RED" "$NC" "$1" >&2; exit 1; }

show_help() {
    cat <<'EOF'
Gno binary uninstaller.

Usage:
  curl --proto '=https' --tlsv1.2 -sSf \
    https://raw.githubusercontent.com/gnolang/gno/master/misc/uninstall.sh | sh

Removes binaries from the install dir, and — for users migrating from the
previous source-build installer — also from $GOPATH/bin and the legacy
~/.gno/src source checkout.

Flags:
  --dir <path>      install directory (default: $HOME/.gno/bin)
  --help            show this help

Environment:
  GNO_INSTALL_DIR   same as --dir
EOF
}

parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --dir)     [ $# -ge 2 ] || die "--dir needs a value"; INSTALL_DIR="$2"; shift 2 ;;
            -h|--help) show_help; exit 0 ;;
            *)         die "unknown flag: $1 (try --help)" ;;
        esac
    done
}

uninstall_gno() {
    log "removing from $INSTALL_DIR"
    for c in $FULL_COMPONENTS; do
        rm -f "$INSTALL_DIR/$c"
    done
    if command -v go >/dev/null 2>&1; then
        gopath="$(go env GOPATH 2>/dev/null)"
        gobin="${gopath%/}/bin"
        if [ -n "$gopath" ] && [ "$gobin" != "/bin" ]; then
            log "removing legacy binaries from $gobin"
            for c in $FULL_COMPONENTS; do
                rm -f "$gobin/$c"
            done
        fi
    fi
    if [ -d "$HOME/.gno/src" ]; then
        log "removing legacy source dir $HOME/.gno/src"
        rm -rf "$HOME/.gno/src"
    fi
    log "uninstalled"
}

main() {
    parse_args "$@"
    uninstall_gno
}

main "$@"
