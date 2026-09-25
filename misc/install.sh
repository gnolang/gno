#!/bin/sh
# Gno installer.
# Default mode: download precompiled binaries (Linux/macOS, amd64/arm64).
# --from-source: clone the repo and build via `make install`.
# Run with --help for usage.

set -eu

REPO="gnolang/gno"
API="https://api.github.com/repos/${REPO}"

COMPONENTS="gno gnokey gnodev gnoweb"
FULL_COMPONENTS="gno gnokey gnodev gnoweb gnoland"

VERSION="${GNO_VERSION:-latest}"
INSTALL_DIR="${GNO_INSTALL_DIR:-${HOME}/.gno/bin}"
SRC_DIR="${GNO_SRC_DIR:-${HOME}/.gno/src}"
FULL=0
FROM_SOURCE=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { printf '%b[gno-install]%b %s\n' "$GREEN"  "$NC" "$1"; }
warn() { printf '%b[gno-install]%b %s\n' "$YELLOW" "$NC" "$1" >&2; }
die()  { printf '%b[gno-install] error:%b %s\n' "$RED" "$NC" "$1" >&2; exit 1; }

show_help() {
    cat <<'EOF'
Gno installer.

Default mode downloads precompiled binaries (Linux/macOS, amd64/arm64).
Use --from-source to clone and build with `make install` instead.

Usage:
  curl --proto '=https' --tlsv1.2 -sSf \
    https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh

Flags:
  --version <tag>   install a specific release tag (default: latest).
                    With --from-source, selects the git ref (default: master).
  --dir <path>      install directory (default: $HOME/.gno/bin).
                    With --from-source, used as GOBIN.
  --full            also install the validator node (gnoland)
  --from-source     clone the repository and build with `make install`
                    instead of downloading prebuilt binaries.
                    Requires go, git, and make.
  --help            show this help

By default installs: gno, gnokey, gnodev, gnoweb.
Use --full to additionally install gnoland (validator node).
To remove an installation, see misc/uninstall.sh.

Environment:
  GNO_VERSION       same as --version
  GNO_INSTALL_DIR   same as --dir
  GNO_SRC_DIR       source checkout dir for --from-source
                    (default: $HOME/.gno/src)
  GITHUB_TOKEN      optional. authenticates GitHub API requests to raise the
                    60 requests/hour anonymous rate limit
EOF
}

parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --version)     [ $# -ge 2 ] || die "--version needs a value"; VERSION="$2"; shift 2 ;;
            --dir)         [ $# -ge 2 ] || die "--dir needs a value"; INSTALL_DIR="$2"; shift 2 ;;
            --full)        FULL=1; shift ;;
            --from-source) FROM_SOURCE=1; shift ;;
            -h|--help)     show_help; exit 0 ;;
            *)             die "unknown flag: $1 (try --help)" ;;
        esac
    done
}

detect_platform() {
    case "$(uname -s)" in
        Linux)  OS="linux" ;;
        Darwin) OS="darwin" ;;
        *) die "unsupported OS: $(uname -s)" ;;
    esac
    case "$(uname -m)" in
        x86_64|amd64)  ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) die "unsupported architecture: $(uname -m)" ;;
    esac
}

check_deps() {
    if   command -v curl >/dev/null 2>&1; then HTTP_TOOL="curl"
    elif command -v wget >/dev/null 2>&1; then HTTP_TOOL="wget"
    else die "curl or wget is required"
    fi
    command -v tar     >/dev/null 2>&1 || die "tar is required"
    command -v install >/dev/null 2>&1 || die "install is required"

    if   command -v sha256sum >/dev/null 2>&1; then SHA="sha256sum"
    elif command -v shasum    >/dev/null 2>&1; then SHA="shasum -a 256"
    else die "sha256sum or shasum is required"

    fi
    if command -v jq >/dev/null 2>&1; then JSON="jq"; else JSON="awk"; fi
}

# Stack-based xtrace suspension; safe across nested callers and subshells.
suspend_xtrace() {
    case "$-" in
        *x*) _xt_stack="1${_xt_stack-}"; set +x ;;
        *)   _xt_stack="0${_xt_stack-}" ;;
    esac
}

restore_xtrace() {
    case "${_xt_stack-}" in
        1*) _xt_stack="${_xt_stack#?}"; set -x ;;
        0*) _xt_stack="${_xt_stack#?}" ;;
        *)  : ;;
    esac
}

# Do not use for asset downloads: asset URLs redirect to another host and
# Authorization headers must not be forwarded to the CDN.
api_get() {
    _headers="$TMP/api_headers"
    if [ "$HTTP_TOOL" = "curl" ]; then
        if [ -n "${GH_API_TOKEN:+x}" ]; then
            suspend_xtrace
            printf 'header = "Authorization: Bearer %s"\n' "$GH_API_TOKEN" \
                | $CURL -D "$_headers" --config - "$@"
            _rc=$?
            restore_xtrace
        else
            $CURL -D "$_headers" "$@"
            _rc=$?
        fi
    else
        # Capture response headers via stderr so the rate-limit check below
        # can scan them. Auth (when present) comes from $WGET_AUTH_CFG so the
        # token never appears on argv (visible via /proc/<pid>/cmdline).
        if [ -n "$WGET_AUTH_CFG" ]; then
            $WGET --config="$WGET_AUTH_CFG" --no-verbose -S -O - "$@" 2>"$_headers"
        else
            $WGET --no-verbose -S -O - "$@" 2>"$_headers"
        fi
        _rc=$?
    fi
    if [ "$_rc" -ne 0 ] && [ -z "${GH_API_TOKEN:+x}" ] && [ -f "$_headers" ] \
        && grep -qi '^[[:space:]]*x-ratelimit-remaining:[[:space:]]*0[[:space:]]*$' "$_headers" 2>/dev/null; then
        log "GitHub API rate limit exhausted (60/hour anonymous)" >&2
        log "set GITHUB_TOKEN to authenticate; see --help" >&2
    fi
    return $_rc
}

# Intentionally does not follow redirects: we need the redirect target
# (signed URL), not the asset content, and Authorization must not reach
# the CDN host.
resolve_asset() {
    if [ "$HTTP_TOOL" = "curl" ]; then
        if [ -n "${GH_API_TOKEN:+x}" ]; then
            suspend_xtrace
            printf 'header = "Authorization: Bearer %s"\n' "$GH_API_TOKEN" \
                | $CURL --config - \
                    -H "Accept: application/octet-stream" \
                    -o /dev/null -w '%{redirect_url}' \
                    "$1"
            _rc=$?
            restore_xtrace
            return $_rc
        fi
        $CURL \
            -H "Accept: application/octet-stream" \
            -o /dev/null -w '%{redirect_url}' \
            "$1"
        return $?
    fi
    # --max-redirect=0 makes the 3xx response a non-zero exit, but -S still
    # prints the Location header. Ignore the exit and parse it ourselves.
    _hdr="$TMP/wget_redir.$$"
    if [ -n "$WGET_AUTH_CFG" ]; then
        $WGET --config="$WGET_AUTH_CFG" --no-verbose --max-redirect=0 -S \
            --header="Accept: application/octet-stream" \
            -O /dev/null "$1" >/dev/null 2>"$_hdr" || true
    else
        $WGET --no-verbose --max-redirect=0 -S \
            --header="Accept: application/octet-stream" \
            -O /dev/null "$1" >/dev/null 2>"$_hdr" || true
    fi
    sed -n 's/\r$//; s/^[[:space:]]*[Ll][Oo][Cc][Aa][Tt][Ii][Oo][Nn]:[[:space:]]*//p' "$_hdr" | head -1
    rm -f "$_hdr"
}

# Move GITHUB_TOKEN into a non-exported variable before spawning children.
capture_github_token() {
    GH_API_TOKEN=
    if [ -n "${GITHUB_TOKEN:+x}" ]; then
        suspend_xtrace
        GH_API_TOKEN=$GITHUB_TOKEN
        unset GITHUB_TOKEN
        restore_xtrace
    fi
}

release_tag() {
    if [ "$JSON" = "jq" ]; then
        jq -r '.tag_name' "$TMP/release.json"
    else
        sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' "$TMP/release.json" | head -1
    fi
}

# /releases/latest may point at a chain/* tag, which pins a network rather than
# naming a version to install, so we walk the list and pick the first
# non-prerelease v* tag instead.
latest_v_tag() {
    if [ "$JSON" = "jq" ]; then
        jq -r 'map(select(.prerelease == false and (.tag_name | startswith("v")))) | .[0].tag_name // empty' "$TMP/releases.json"
    else
        # GitHub lists releases newest-first. Pair each v-prefixed tag_name with
        # the prerelease flag that follows it in the same release object, and
        # emit the first non-prerelease one.
        awk '
            /"tag_name":/ {
                line = $0
                sub(/.*"tag_name"[[:space:]]*:[[:space:]]*"/, "", line)
                sub(/".*/, "", line)
                if (line ~ /^v/) tag = line; else tag = ""
                next
            }
            /"prerelease":/ {
                if (tag != "" && $0 !~ /true/) { print tag; exit }
                tag = ""
            }
        ' "$TMP/releases.json"
    fi
}

# The awk fallback scans pretty-printed JSON and relies on the current
# field order ("url" before "name" within each asset).
asset_url() {
    if [ "$JSON" = "jq" ]; then
        jq -r --arg n "$1" '.assets[] | select(.name == $n) | .url' "$TMP/release.json"
    else
        awk -v t="$1" -v api="$API" '
            match($0, /releases\/assets\/[0-9]+/) { u = substr($0, RSTART, RLENGTH) }
            /"name":/ && index($0, "\"" t "\"") && u { print api "/" u; exit }
        ' "$TMP/release.json"
    fi
}

# Plain GET to file, following redirects. No auth (signed CDN URLs only).
http_get() {
    _out="$1"; _url="$2"
    if [ "$HTTP_TOOL" = "curl" ]; then
        $CURL -L -o "$_out" "$_url"
    else
        $WGET -q -O "$_out" "$_url"
    fi
}

install_gno() {
    # curl: --proto =https and --tlsv1.2 harden the transport, --retry handles flaky nets.
    # -L is added per-call: resolve_asset needs the redirect URL (no follow),
    # downloads of signed URLs follow redirects explicitly.
    # wget: --https-only enforces HTTPS for redirects; retry/waitretry mirror curl.
    CURL="curl --proto =https --tlsv1.2 -fsS --retry 3 --retry-delay 2"
    WGET="wget --https-only --tries=3 --waitretry=2"

    [ "$HTTP_TOOL" = "wget" ] && log "curl not found; using wget for downloads"

    if [ -n "${GH_API_TOKEN:+x}" ]; then
        log "authenticating GitHub API requests with GITHUB_TOKEN"
    fi

    TMP="$(mktemp -d)"
    trap 'rm -rf "$TMP"' EXIT INT TERM

    # wget cannot read a config from stdin like curl can. Stage the bearer
    # token in a 0600 file once; the trap above wipes it on exit/interrupt.
    WGET_AUTH_CFG=""
    if [ "$HTTP_TOOL" = "wget" ] && [ -n "${GH_API_TOKEN:+x}" ]; then
        WGET_AUTH_CFG="$TMP/wget_auth.cfg"
        suspend_xtrace
        ( umask 077; printf 'header = Authorization: Bearer %s\n' "$GH_API_TOKEN" >"$WGET_AUTH_CFG" )
        restore_xtrace
    fi

    # GitHub's /releases/latest resolves to whatever it ranks as "latest",
    # which for this repo may be a chain/* tag. Resolve "latest" to the most
    # recent non-prerelease v* tag ourselves instead.
    if [ "$VERSION" = "latest" ]; then
        # per_page is the whole search window: this walks one page and does not
        # paginate. chain/* tags are cut far more often than v* ones, so a small
        # window eventually holds nothing but chain tags and "latest" stops
        # resolving. 100 is the API maximum.
        api_get "${API}/releases?per_page=100" > "$TMP/releases.json" \
            || die "failed to fetch releases list"
        VERSION="$(latest_v_tag)"
        [ -n "$VERSION" ] || die "no v* release found; pass --version <tag> explicitly (see https://github.com/${REPO}/releases)"
    fi

    # The public github.com/<repo>/releases/download/... path currently 404s for
    # this repository; resolving assets via the API endpoint works around it.
    META_URL="${API}/releases/tags/${VERSION}"
    if ! api_get "$META_URL" > "$TMP/release.json" 2>"$TMP/meta_err"; then
        # A tag can exist in git with no Release object behind it: the release
        # workflow is what creates the Release, and until it has run there is
        # nothing to download. That is a different failure from a typo'd tag and
        # it has a way through, so do not report both as "metadata fetch failed".
        if api_get "${API}/git/ref/tags/${VERSION}" >/dev/null 2>&1; then
            die "$VERSION is a git tag but has no published release, so there are no binaries to download. Build that exact ref instead: --from-source --version $VERSION"
        fi
        # Any other cause: replay what the API said, including the rate-limit
        # hint api_get emits, before giving up on it.
        cat "$TMP/meta_err" >&2
        die "no release tagged $VERSION (see https://github.com/${REPO}/releases)"
    fi

    VERSION="$(release_tag)"
    [ -n "$VERSION" ] || die "could not parse tag_name from release metadata"

    log "installing gno ${VERSION} (${OS}/${ARCH}) into ${INSTALL_DIR}"

    if [ "$FULL" = 1 ]; then
        components="$FULL_COMPONENTS"
    else
        components="$COMPONENTS"
    fi

    # Two release shapes exist, and "latest" can resolve to either: the older
    # goreleaser releases ship one tarball per platform, release-chain-tag.yml
    # ships one binary per target plus CHECKSUMS.txt. Pick by what the release
    # actually carries.
    ARCHIVE="gno_${VERSION#v}_${OS}_${ARCH}.tar.gz"
    ARCHIVE_URL="$(asset_url "$ARCHIVE")"
    if [ -n "$ARCHIVE_URL" ]; then
        install_from_archive
    elif [ -n "$(asset_url "gno_${OS}_${ARCH}")" ]; then
        install_from_binaries
    else
        die "$VERSION has no binaries for ${OS}/${ARCH} (looked for $ARCHIVE and gno_${OS}_${ARCH})"
    fi

    log "installed into $INSTALL_DIR"
    print_next_steps
}

# goreleaser's shape: one tar.gz per platform, plus lowercase checksums.txt.
install_from_archive() {
    SUMS_URL="$(asset_url "checksums.txt")"
    [ -n "$SUMS_URL" ] || die "checksums.txt missing from $VERSION"

    log "downloading $ARCHIVE"
    # Signed CDN URLs carry short-lived query credentials; keep them out of xtrace.
    suspend_xtrace
    ARCHIVE_SIGNED="$(resolve_asset "$ARCHIVE_URL")"
    [ -n "$ARCHIVE_SIGNED" ] || die "could not resolve $ARCHIVE download URL"
    http_get "$TMP/$ARCHIVE"      "$ARCHIVE_SIGNED" || die "archive download failed"

    SUMS_SIGNED="$(resolve_asset "$SUMS_URL")"
    [ -n "$SUMS_SIGNED" ] || die "could not resolve checksums.txt download URL"
    http_get "$TMP/checksums.txt" "$SUMS_SIGNED"    || die "checksums download failed"
    restore_xtrace

    expected="$(awk -v n="$ARCHIVE" '$2 == n {print $1; exit}' "$TMP/checksums.txt")"
    [ -n "$expected" ] || die "$ARCHIVE not listed in checksums.txt"
    actual="$(cd "$TMP" && $SHA "$ARCHIVE" | awk '{print $1}')"
    [ "$expected" = "$actual" ] || die "sha256 mismatch: expected $expected, got $actual"
    log "sha256 verified"

    mkdir -p "$INSTALL_DIR" "$TMP/ext"
    tar -xzf "$TMP/$ARCHIVE" -C "$TMP/ext"
    missing=""
    installed_count=0
    for c in $components; do
        if [ ! -f "$TMP/ext/$c" ]; then
            missing="${missing} ${c}"
            continue
        fi
        install_binary "$TMP/ext/$c" "$c"
        installed_count=$((installed_count + 1))
    done
    [ "$installed_count" -gt 0 ] || die "no expected binaries found in $ARCHIVE (missing:${missing})"
    [ -z "$missing" ] || log "warning: expected binaries missing from $ARCHIVE:${missing}"
}

# release / chain-tag's shape: <name>_<os>_<arch> per binary, plus an uppercase
# CHECKSUMS.txt covering all platforms. It ships the chain binaries and gnodev
# only, so anything else is reported as missing, not fatal.
install_from_binaries() {
    SUMS_URL="$(asset_url "CHECKSUMS.txt")"
    [ -n "$SUMS_URL" ] || die "CHECKSUMS.txt missing from $VERSION"

    suspend_xtrace
    SUMS_SIGNED="$(resolve_asset "$SUMS_URL")"
    [ -n "$SUMS_SIGNED" ] || die "could not resolve CHECKSUMS.txt download URL"
    http_get "$TMP/CHECKSUMS.txt" "$SUMS_SIGNED" || die "checksums download failed"
    restore_xtrace

    mkdir -p "$INSTALL_DIR"
    missing=""
    installed_count=0
    for c in $components; do
        name="${c}_${OS}_${ARCH}"
        url="$(asset_url "$name")"
        if [ -z "$url" ]; then
            missing="${missing} ${c}"
            continue
        fi

        log "downloading $name"
        suspend_xtrace
        signed="$(resolve_asset "$url")"
        [ -n "$signed" ] || die "could not resolve $name download URL"
        http_get "$TMP/$name" "$signed" || die "$name download failed"
        restore_xtrace

        # Tolerate a leading "./": shasum writes the path it was given.
        expected="$(awk -v n="$name" '{ f = $2; sub(/^\.\//, "", f); if (f == n) { print $1; exit } }' "$TMP/CHECKSUMS.txt")"
        [ -n "$expected" ] || die "$name not listed in CHECKSUMS.txt"
        actual="$(cd "$TMP" && $SHA "$name" | awk '{print $1}')"
        [ "$expected" = "$actual" ] || die "sha256 mismatch for $name: expected $expected, got $actual"

        install_binary "$TMP/$name" "$c"
        installed_count=$((installed_count + 1))
    done
    [ "$installed_count" -gt 0 ] || die "no expected binaries found in $VERSION (missing:${missing})"
    [ -z "$missing" ] || log "warning: not published in $VERSION:${missing}"
    log "sha256 verified"
}

install_binary() {
    install -m 0755 "$1" "$INSTALL_DIR/$2"
    # Best-effort Gatekeeper unblock on macOS; harmless on Linux.
    [ "$OS" = "darwin" ] && xattr -d com.apple.quarantine "$INSTALL_DIR/$2" 2>/dev/null || true
    return 0
}

print_next_steps() {
    [ -x "$INSTALL_DIR/gno" ] || die "installation failed: $INSTALL_DIR/gno not found"
    # Run it rather than trust the download: a binary can install cleanly and
    # still be unusable. Releases built with -trimpath produce a gno that cannot
    # work out its own GNOROOT and panics on every subcommand, which looks
    # identical on the release page. Name that case instead of guessing at
    # architecture, which is what the old message did and it was always wrong.
    #
    # Run it from / rather than here: gnoenv also resolves GNOROOT by asking
    # `go list` about the module of the current directory, so a binary that
    # cannot find its stdlib on its own still passes this check when the
    # installer happens to be run from inside a gno checkout, which is exactly
    # where CI runs it. --dir accepts a relative path, so resolve it first.
    gno_bin="$INSTALL_DIR/gno"
    case "$gno_bin" in /*) ;; *) gno_bin="$(pwd)/$gno_bin" ;; esac
    if ! gno_check="$(cd / && "$gno_bin" version 2>&1)"; then
        printf '%s\n' "$gno_check" >&2
        case "$gno_check" in
            *GNOROOT*) warn "gno cannot locate its standard library. These binaries were built without a usable GNOROOT; install a newer release (--version latest) or build locally (--from-source)." ;;
            *)         warn "gno installed but failed to run; the binary may not be compatible with this system" ;;
        esac
    else
        printf '%s\n' "$gno_check"
    fi

    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *) log "add to PATH: export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
    esac

    cat <<'EOF'

                    __             __
  ___ ____  ___    / /__ ____  ___/ /
 / _ `/ _ \/ _ \_ / / _ `/ _ \/ _  /
 \_, /_//_/\___(_)_/\_,_/_//_/\_,_//
/___/

 To get started:
    gnodev                      Run a local chain with hot reload
    https://gno.land            Explore realms already deployed on-chain
    https://docs.gno.land       Full guide + deploy to a live network

EOF
}

install_from_source() {
    command -v go   >/dev/null 2>&1 || die "go is not installed. See https://go.dev/doc/install"
    command -v git  >/dev/null 2>&1 || die "git is not installed. See https://git-scm.com/downloads"
    command -v make >/dev/null 2>&1 || die "make is not installed; install your platform's build tools"

    if [ "$VERSION" = "latest" ]; then
        ref="master"
    else
        ref="$VERSION"
    fi

    # Shallow clone to keep disk + network cost low (matches the previous
    # installer's behavior). reset --hard FETCH_HEAD on re-runs ensures
    # a stale local branch does not get reinstalled.
    if [ ! -d "$SRC_DIR/.git" ]; then
        log "cloning ${REPO} ($ref) into $SRC_DIR"
        mkdir -p "$SRC_DIR"
        git clone --depth 1 --branch "$ref" --quiet \
            "https://github.com/${REPO}.git" "$SRC_DIR" || die "git clone $ref failed"
    else
        log "updating $SRC_DIR to $ref"
        git -C "$SRC_DIR" fetch --depth 1 --quiet origin "$ref" || die "git fetch $ref failed"
        git -C "$SRC_DIR" reset --hard --quiet FETCH_HEAD || die "git reset $ref failed"
    fi

    mkdir -p "$INSTALL_DIR"
    log "building from source (ref=$ref, GOBIN=$INSTALL_DIR)"
    GOBIN="$INSTALL_DIR" make --no-print-directory -C "$SRC_DIR" install install.gnobro \
        || die "make install failed"
    GOBIN="$INSTALL_DIR" make --no-print-directory -C "$SRC_DIR/gno.land" install.gnoweb \
        || die "make install.gnoweb failed"
    if [ "$FULL" = 1 ]; then
        GOBIN="$INSTALL_DIR" make --no-print-directory -C "$SRC_DIR/gno.land" install.gnoland \
            || die "make install.gnoland failed"
    fi

    log "installed into $INSTALL_DIR"
    print_next_steps
}

main() {
    parse_args "$@"
    if [ "$FROM_SOURCE" = 1 ]; then
        install_from_source
    else
        capture_github_token
        detect_platform
        check_deps
        install_gno
    fi
}

main "$@"
