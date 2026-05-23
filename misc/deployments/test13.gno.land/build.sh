#!/usr/bin/env bash
# build.sh — master wrapper for the test-13 hardfork build pipeline.
#
# Runs phase-1-build-genesis.sh, then phase-2-apply-replay.sh. Forwards
# any args verbatim to both phase scripts; each phase ignores flags it
# doesn't recognise via its own argument parser (so e.g. --skip-audit
# is consumed by phase-2 and rejected by phase-1 — pass it directly to
# phase-2 if you want to use it).
#
# To keep the pipeline simple, this wrapper only accepts the union of
# flags both scripts share: --debug, --no-install. Phase-2-only flags
# (--skip-audit, --source-txs-*) should be passed by running phase-2
# directly after phase-1 has completed once.
#
# Usage:
#   ./build.sh                  # full pipeline
#   ./build.sh --debug          # show every command being run
#   ./build.sh --no-install     # reuse previously built binaries
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
. "$SCRIPT_DIR/lib/common.sh"

# Validate args before kicking off — phase-1 and phase-2 use different
# argument parsers and would error on unknown flags mid-run. Catching
# typos here saves ~60s of binary builds before the fail.
for arg in "$@"; do
  case "$arg" in
  --debug | --no-install) ;;
  *)
    printf 'ERROR: build.sh accepts only --debug and --no-install.\n' >&2
    printf '       For phase-2-only flags (--skip-audit, --source-txs-*),\n' >&2
    printf '       run ./phase-1-build-genesis.sh then ./phase-2-apply-replay.sh directly.\n' >&2
    printf '       Unknown argument: %s\n' "$arg" >&2
    exit 1
    ;;
  esac
done

BUILD_START_TS=$(date +%s)

printf '\n### test-13 build pipeline ###\n'

"$SCRIPT_DIR/phase-1-build-genesis.sh" "$@"

printf '\n### Phase 1 -> Phase 2 handoff ###\n'

"$SCRIPT_DIR/phase-2-apply-replay.sh" "$@"

BUILD_END_TS=$(date +%s)
BUILD_DURATION=$((BUILD_END_TS - BUILD_START_TS))

FINAL_GENESIS="$SCRIPT_DIR/genesis.json"
if [ -f "$FINAL_GENESIS" ]; then
  FINAL_SHA=$(sha256_of "$FINAL_GENESIS")
  FINAL_BYTES=$(file_size "$FINAL_GENESIS")
  printf '\n### test-13 build complete: genesis.json (%s, sha256=%s) ###\n' \
    "$(format_size "$FINAL_BYTES")" "$FINAL_SHA"
  printf '    total pipeline time: %s\n' "$(format_duration "$BUILD_DURATION")"
else
  printf '\n### test-13 build complete (pipeline ran but %s was not produced) ###\n' \
    "$FINAL_GENESIS"
  printf '    total pipeline time: %s\n' "$(format_duration "$BUILD_DURATION")"
  exit 1
fi
