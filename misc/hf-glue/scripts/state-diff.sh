#!/usr/bin/env bash
# state-diff.sh — diff realm render output between the source chain at
# halt_height and our post-replay node. Catches silent state divergence
# introduced by the 2605 "Unable to deliver genesis tx" failures that
# --skip-failing-genesis-txs absorbs.
#
# Approach
# ========
#   For each realm in REALMS (one pkgpath[:subpath] per line), fetch
#   vm/qrender output from both sides:
#     source chain: `gnokey query -remote $SOURCE_RPC -height $HALT_HEIGHT`
#     replay:       `gnokey query -remote $REPLAY_RPC`
#   Normalize out obviously-ephemeral bits (absolute paths, wall clock,
#   SVG blobs, current block height, etc.), then byte-compare. Emit a
#   STATE-DIFF.md with per-realm pass/fail; for failing realms include the
#   first N lines of the unified diff.
#
# What this can't catch
# =====================
#   A realm's Render function only surfaces what it chooses to. State
#   stored in maps the render ignores, private fields, etc., are invisible.
#   For those realms, add a companion qeval check in a follow-up.
#
# Env
# ===
#   SOURCE_RPC     source chain RPC (default https://rpc.gno.land)
#   REPLAY_RPC     post-replay node RPC (default http://localhost:26657)
#   HALT_HEIGHT    source-chain height to query (required; same value used
#                  at genesis build time)
#   REALMS         newline-separated pkgpath[:subpath] list; defaults to
#                  the canonical set (valset v2, govDAO, memberstore,
#                  users, home, blog, sys/params, sys/names)
#   OUT            misc/hf-glue/out (auto-resolved)
#   DIFF_CONTEXT   lines of context per failing diff (default 20)
#   GNOKEY_BIN     gnokey binary (default: gnokey on $PATH)
#
# Exit status
# ===========
#   0  — every realm matches after normalization
#   1  — at least one realm diverges (see STATE-DIFF.md)
#   2  — prerequisite error
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="${OUT:-$(cd "$SCRIPT_DIR/.." && pwd)/out}"
mkdir -p "$OUT"

SOURCE_RPC="${SOURCE_RPC:-https://rpc.gno.land}"
REPLAY_RPC="${REPLAY_RPC:-http://localhost:26657}"
DIFF_CONTEXT="${DIFF_CONTEXT:-20}"
GNOKEY_BIN="${GNOKEY_BIN:-gnokey}"

: "${HALT_HEIGHT:?HALT_HEIGHT is required (pin the source-chain height to query)}"

# The replay node can be queried at historical heights (0 = current tip).
# Production recipe: run this script against a FRESH post-replay node
# before any on-chain activity can drift the state — the default tip
# matches initial_height in that case. Override REPLAY_HEIGHT to pin a
# specific historical height; requires the node to still retain state at
# that height (no pruning).
REPLAY_HEIGHT="${REPLAY_HEIGHT:-0}"

# Default realm set. Each entry is a pkgpath with optional :subpath passed
# to vm/qrender. Extend via REALMS env (newline-separated) for ad-hoc runs.
REALMS="${REALMS:-$(
  cat <<'EOF'
gno.land/r/sys/validators/v2:
gno.land/r/sys/names:
gno.land/r/sys/params:
gno.land/r/gov/dao:
gno.land/r/gov/dao/v3/memberstore:members
gno.land/r/gnoland/home:
gno.land/r/gnoland/blog:
gno.land/r/gnoland/users:
EOF
)}"

command -v "$GNOKEY_BIN" >/dev/null 2>&1 || {
  echo "gnokey not found on PATH (set GNOKEY_BIN=...)" >&2
  exit 2
}
command -v diff >/dev/null 2>&1 || {
  echo "diff not found on PATH" >&2
  exit 2
}

REPORT="$OUT/STATE-DIFF.md"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# ---- Normalization
# Ephemeral render bits that differ between mainnet and our replay even for
# logically-equivalent state:
#   • embedded SVG charts (base64 blobs with sizes in px → drop wholesale)
#   • absolute URLs pointing at the chain's own rpc/web (host differs)
#   • wall-clock timestamps ("Generated at ...")
#   • "height" / "block" markers that reference the chain's current tip
#   • trailing whitespace
# The goal is to preserve semantically-meaningful text (addresses, names,
# ids, counts, titles, descriptions) and drop everything transient.
normalize() {
  sed \
    -e '/data:image\/svg+xml;base64,/d' \
    -e 's|https://[^[:space:])]*|<URL>|g' \
    -e 's|http://[^[:space:])]*|<URL>|g' \
    -e 's|rpc\.gno\.land|<HOST>|g' \
    -e 's|latest_block_height[^[:space:]]*|<BLOCK>|g' \
    -e 's|height=[0-9]*|height=<H>|g' \
    -e 's|Generated [^\\n]*|Generated <TIME>|g' \
    -e 's|^\([[:space:]]*[0-9]*[[:space:]]*\)/src/|\1<SRC>/|g' \
    -e 's|^\([[:space:]]*[0-9]*[[:space:]]*\)gno/|\1<SRC>/|g' \
    -e 's|^\([[:space:]]*[0-9]*[[:space:]]*\)/gnoroot/|\1<SRC>/|g' \
    -e 's|^\([[:space:]]*[0-9]*[[:space:]]*\)/usr/lib/go/src/|\1<GOROOT>/|g' \
    -e 's|^\([[:space:]]*[0-9]*[[:space:]]*\)/usr/local/go/src/|\1<GOROOT>/|g' \
    -e 's|^\([[:space:]]*[0-9]*[[:space:]]*\)/usr/pkg/mod/|\1<GOMOD>/|g' \
    -e 's|^\([[:space:]]*[0-9]*[[:space:]]*\)/root/\.cache/go-build/|\1<GOMOD>/|g' \
    -e 's|errors\.go:[0-9]*|errors.go:<L>|g' \
    -e 's|keeper\.go:[0-9]*|keeper.go:<L>|g' \
    -e 's|handler\.go:[0-9]*|handler.go:<L>|g' \
    -e 's|baseapp\.go:[0-9]*|baseapp.go:<L>|g' \
    -e 's|local_client\.go:[0-9]*|local_client.go:<L>|g' \
    -e 's|app_conn\.go:[0-9]*|app_conn.go:<L>|g' \
    -e 's|abci\.go:[0-9]*|abci.go:<L>|g' \
    -e 's|http_server\.go:[0-9]*|http_server.go:<L>|g' \
    -e 's|handlers\.go:[0-9]*|handlers.go:<L>|g' \
    -e 's|server\.go:[0-9]*|server.go:<L>|g' \
    -e 's|asm[^.]*\.s:[0-9]*|asm.s:<L>|g' \
    -e 's|reflect/value\.go:[0-9]*|reflect/value.go:<L>|g' \
    -e 's|[[:space:]]*$||'
}

query_render() {
  local rpc="$1" realm="$2" height="${3:-0}"
  local args=(-remote "$rpc")
  if [[ "$height" -gt 0 ]]; then
    args+=(-height "$height")
  fi
  # gnokey prints `height: X\ndata: <render output>\n`; drop the header
  # line and keep the body.
  "$GNOKEY_BIN" query "${args[@]}" vm/qrender --data "$realm" 2>/dev/null |
    sed '1,/^data: /s/^data: //; 1d' || true
}

# ---- Report header
{
  echo "# State reconciliation report"
  echo ""
  echo "_Generated $(date -u +%Y-%m-%dT%H:%M:%SZ)_"
  echo ""
  echo "- **Source**: \`$SOURCE_RPC\` at \`height=$HALT_HEIGHT\`"
  echo "- **Replay**: \`$REPLAY_RPC\` at current tip"
  echo ""
  echo "Each realm's rendered output is queried from both sides, normalized"
  echo "to strip ephemeral bits (timestamps, SVGs, URLs, current heights),"
  echo "and byte-compared. A PASS means the source and replay agree on"
  echo "everything the realm's Render function chose to expose. A FAIL"
  echo "means the post-replay state is visibly diverged from source — or"
  echo "the realm exposes a truly ephemeral field we failed to normalize."
  echo ""
  echo "| Realm | Result |"
  echo "|-------|--------|"
} >"$REPORT"

summary_rows=()
fail_sections=()
failed=0
passed=0

# ---- Per-realm diff
while IFS= read -r realm; do
  [[ -z "$realm" ]] && continue
  src="$WORK/src.$(echo "$realm" | tr -c '[:alnum:]' _)"
  rep="$WORK/rep.$(echo "$realm" | tr -c '[:alnum:]' _)"

  query_render "$SOURCE_RPC" "$realm" "$HALT_HEIGHT" | normalize >"$src"
  query_render "$REPLAY_RPC" "$realm" "$REPLAY_HEIGHT" | normalize >"$rep"

  if diff -q "$src" "$rep" >/dev/null 2>&1; then
    summary_rows+=("| \`$realm\` | ✅ PASS |")
    passed=$((passed + 1))
  else
    summary_rows+=("| \`$realm\` | ❌ FAIL ([see below](#$(printf '%s' "$realm" | tr -c '[:alnum:]' '-'))) |")
    # Capture first DIFF_CONTEXT lines of unified diff for triage.
    {
      echo ""
      echo "### $realm"
      echo ""
      echo '```diff'
      # `diff` exits 1 when files differ and `head` closes early; under
      # set -euo pipefail that would kill the script mid-loop.
      diff -u "$src" "$rep" | head -n "$DIFF_CONTEXT" || true
      echo '```'
    } >>"$WORK/fail_details"
    failed=$((failed + 1))
  fi
done <<<"$REALMS"

{
  for row in "${summary_rows[@]}"; do echo "$row"; done
  echo ""
  echo "- passed: **$passed**"
  echo "- failed: **$failed**"
  if [[ $failed -gt 0 ]]; then
    echo ""
    echo "---"
    echo ""
    echo "## Divergences"
    cat "$WORK/fail_details"
  fi
} >>"$REPORT"

echo "State diff report: $REPORT"
echo ""
printf 'Summary: %d passed, %d failed\n' "$passed" "$failed"
exit "$failed"
