#!/usr/bin/env bash
# common.sh — shared helpers for the test-13 hardfork build pipeline.
#
# Sourced by phase-1-build-genesis.sh, phase-2-apply-replay.sh, and build.sh.
# Must remain bash 3.2-compatible (default macOS) — no associative arrays,
# no `mapfile`, no `${var,,}` lowercasing, no `&&`/`||` inside `[[ ]]`.
#
# Exposes:
#   download_url <url> <dest>          curl → wget fallback
#   sha256_of <path>                   shasum -a 256 → sha256sum fallback (lowercase hex)
#   gunzip_to_stdout <gz_path>         gunzip -c → gzip -dc fallback
#   require_tools <tool> [<tool>...]   batch preflight with per-tool install hints
#   verify_checksum <path>             match against CHECKSUMS.txt at TEST13_DIR root
#   print_phase_header <phase> <step> <total> <title>
#   print_substep <code> <text>
#   die <msg>                          stderr ERROR: + exit 1
#   format_duration <seconds>          "4 minutes 12 seconds"
#   format_size <bytes>                "245 MB" / "4 KB"
#
# Expects the sourcing script to set TEST13_DIR before calling
# verify_checksum (the test13.gno.land directory used as the relative-path
# root for CHECKSUMS.txt lookups).

# ---- Fatal error reporter

die() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

# ---- Tool dispatchers
# Inline `command -v X` checks in the phase scripts are forbidden — use these.

# download_url <url> <dest_path>
# Tries curl, then wget. Both with sane flags. Errors if neither is available.
download_url() {
  local url="$1"
  local dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$dest" "$url"
  else
    die "neither curl nor wget is installed (need one of them to download $url)"
  fi
}

# sha256_of <path>
# Prints lowercase hex sha256 of the file's content. Tries shasum (macOS +
# most Linux), falls back to sha256sum (some Linux distros without shasum).
sha256_of() {
  local path="$1"
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$path" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$path" | awk '{print $1}'
  else
    die "neither shasum nor sha256sum is installed (need one of them)"
  fi
}

# gunzip_to_stdout <gz_path>
# Prints the decompressed content to stdout. Tries gunzip -c, falls back to
# gzip -dc. Errors if neither is available.
gunzip_to_stdout() {
  local gz="$1"
  if command -v gunzip >/dev/null 2>&1; then
    gunzip -c "$gz"
  elif command -v gzip >/dev/null 2>&1; then
    gzip -dc "$gz"
  else
    die "neither gunzip nor gzip is installed (need one of them to decompress $gz)"
  fi
}

# ---- Tool preflight
# require_tools <tool>...
# Probes every named tool; if any are missing, prints the full list with
# install hints (apt + brew) and exits. Used at the top of each phase
# script so the operator sees all missing tools at once instead of failing
# on the first occurrence mid-run.
#
# Recognised tools (each with a known install hint):
#   curl, wget   — at least one needed (download_url tries both)
#   shasum, sha256sum — at least one needed (sha256_of tries both)
#   gunzip, gzip — at least one needed (gunzip_to_stdout tries both)
#   jq           — required (no fallback)
#   awk, sed, grep, sort, tr, mv, cp, ls, find, wc, head, tail, cut — POSIX core
#   tar          — used by gnogenesis dependencies
#   go           — required to build the gno binaries
#   python3      — used by pick_free_port helper in phase-1
#
# Logical groups (at-least-one): pass any of the alternatives — the function
# treats them as alternatives only when ALL are missing. Otherwise each name
# is checked independently.
require_tools() {
  local missing=""
  local tool
  for tool in "$@"; do
    case "$tool" in
    "curl|wget")
      if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        missing="$missing curl|wget"
      fi
      ;;
    "shasum|sha256sum")
      if ! command -v shasum >/dev/null 2>&1 && ! command -v sha256sum >/dev/null 2>&1; then
        missing="$missing shasum|sha256sum"
      fi
      ;;
    "gunzip|gzip")
      if ! command -v gunzip >/dev/null 2>&1 && ! command -v gzip >/dev/null 2>&1; then
        missing="$missing gunzip|gzip"
      fi
      ;;
    *)
      if ! command -v "$tool" >/dev/null 2>&1; then
        missing="$missing $tool"
      fi
      ;;
    esac
  done

  if [ -z "$missing" ]; then
    return 0
  fi

  printf 'ERROR: missing required tools:\n' >&2
  local m
  for m in $missing; do
    printf '  - %s\n' "$m" >&2
    case "$m" in
    "curl|wget")
      printf '      install:  brew install curl   |   apt-get install -y curl\n' >&2
      ;;
    "shasum|sha256sum")
      printf '      install:  brew install coreutils   |   apt-get install -y coreutils\n' >&2
      ;;
    "gunzip|gzip")
      printf '      install:  brew install gzip   |   apt-get install -y gzip\n' >&2
      ;;
    jq)
      printf '      install:  brew install jq   |   apt-get install -y jq\n' >&2
      ;;
    go)
      printf '      install:  brew install go   |   see https://go.dev/doc/install\n' >&2
      ;;
    python3)
      printf '      install:  brew install python3   |   apt-get install -y python3\n' >&2
      ;;
    awk | sed | grep | sort | tr | mv | cp | ls | find | wc | head | tail | cut | tar)
      printf '      install:  comes with any POSIX userland (coreutils + findutils + tar)\n' >&2
      ;;
    *)
      printf '      install:  consult your package manager\n' >&2
      ;;
    esac
  done
  exit 1
}

# ---- Checksum verification
# verify_checksum <path>
#
# Computes sha256 of the file at <path>, looks up <path> (relative to
# TEST13_DIR) in TEST13_DIR/CHECKSUMS.txt, and one of:
#   - hash matches               → silent OK
#   - hash differs               → FAIL with expected vs got
#   - path not listed            → print computed sha256 + the line to append
#
# CHECKSUMS.txt format: `<sha256>  <relative-path>` (two spaces, matching
# `shasum -a 256` output). Blank lines and `#`-prefixed lines ignored.
verify_checksum() {
  local path="$1"
  if [ -z "${TEST13_DIR:-}" ]; then
    die "verify_checksum: TEST13_DIR not set"
  fi
  if [ ! -f "$path" ]; then
    die "verify_checksum: $path does not exist"
  fi

  local checksums="$TEST13_DIR/CHECKSUMS.txt"
  local rel="${path#"$TEST13_DIR"/}"
  local got
  got=$(sha256_of "$path")

  if [ ! -f "$checksums" ]; then
    printf '  [checksum] %s\n' "$rel" >&2
    printf '             not listed in CHECKSUMS.txt (file missing). Append to lock:\n' >&2
    printf '             %s  %s\n' "$got" "$rel" >&2
    return 0
  fi

  # Find the line whose second field equals $rel. awk is the cleanest
  # bash 3.2-portable way: it splits on whitespace and is available
  # everywhere we run.
  local expected
  expected=$(awk -v rel="$rel" '
    /^[[:space:]]*$/ { next }
    /^[[:space:]]*#/ { next }
    {
      # field 1 is the hash, fields 2..NF re-joined are the path (paths
      # may contain spaces, but ours dont — kept simple)
      if ($2 == rel) { print $1; exit }
    }
  ' "$checksums")

  if [ -z "$expected" ]; then
    printf '  [checksum] %s\n' "$rel" >&2
    printf '             not listed in CHECKSUMS.txt. Append to lock:\n' >&2
    printf '             %s  %s\n' "$got" "$rel" >&2
    return 0
  fi

  if [ "$expected" = "$got" ]; then
    return 0
  fi

  printf 'ERROR: checksum mismatch for %s\n' "$rel" >&2
  printf '       expected: %s\n' "$expected" >&2
  printf '       got:      %s\n' "$got" >&2
  exit 1
}

# ---- Output helpers
# print_phase_header <phase> <step> <total> <title>
#   prints e.g. `=== Phase 1 / Step 3 / 9: Resolve script paths and tooling ===`
print_phase_header() {
  local phase="$1"
  local step="$2"
  local total="$3"
  local title="$4"
  printf '\n=== Phase %s / Step %s of %s: %s ===\n' \
    "$phase" "$step" "$total" "$title"
}

# print_substep <code> <text>
#   prints e.g. `  [1.3.1] Checking cache: work/phase-1/airdrop_balances.txt.gz`
print_substep() {
  local code="$1"
  shift
  printf '  [%s] %s\n' "$code" "$*"
}

# ---- Formatting helpers

# format_duration <seconds>
# Prints "<H> hours <M> minutes <S> seconds" with zero parts omitted.
format_duration() {
  local s="$1"
  if [ "$s" -lt 0 ]; then s=0; fi
  local h=$((s / 3600))
  local m=$(((s % 3600) / 60))
  local sec=$((s % 60))
  local out=""
  if [ "$h" -gt 0 ]; then out="$h hours"; fi
  if [ "$m" -gt 0 ]; then
    if [ -n "$out" ]; then out="$out "; fi
    out="${out}$m minutes"
  fi
  if [ "$sec" -gt 0 ] || [ -z "$out" ]; then
    if [ -n "$out" ]; then out="$out "; fi
    out="${out}$sec seconds"
  fi
  printf '%s' "$out"
}

# format_size <bytes>
# Prints "245 MB", "4 KB", "789 B". Decimal units (1000-based) — matches
# du output on macOS with BLOCKSIZE=1000. Avoids parsing `du -h`, which
# differs across platforms.
format_size() {
  local b="$1"
  if [ "$b" -ge 1000000000 ]; then
    awk -v b="$b" 'BEGIN { printf "%.1f GB", b/1000000000 }'
  elif [ "$b" -ge 1000000 ]; then
    awk -v b="$b" 'BEGIN { printf "%.0f MB", b/1000000 }'
  elif [ "$b" -ge 1000 ]; then
    awk -v b="$b" 'BEGIN { printf "%.0f KB", b/1000 }'
  else
    printf '%s B' "$b"
  fi
}

# file_size <path>  →  bytes (uses wc -c, which is portable; stat flags differ)
file_size() {
  wc -c <"$1" | tr -d ' '
}
