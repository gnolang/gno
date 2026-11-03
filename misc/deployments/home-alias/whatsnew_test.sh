#!/usr/bin/env bash
#
# Tests for whatsnew.sh, offline: the indexer, the RPC and curl are stubbed.
# Usage: bash whatsnew_test.sh
set -uo pipefail

here=$(cd "$(dirname "$0")" && pwd)
script=${WHATSNEW:-$here/whatsnew.sh}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
failed=0

pass() { printf 'ok   %s\n' "$1"; }
fail() {
  printf 'FAIL %s\n' "$1"
  failed=1
}
check() { if eval "$2"; then pass "$1"; else fail "$1"; fi; }

# Each test runs in a subshell that sources the script, so stubs stay local.
load() {
  # shellcheck source=whatsnew.sh
  source "$script"
  set +e
  STATE_DIR=$(mktemp -d "$work/state.XXXXXX")
}

# tx-indexer answers null for a band with no match: an empty band, not an error.
(
  load
  graphql() { echo '{"data":{"getTransactions":null}}'; }
  out=$(scan_band 0 100 'MsgCall: {}' 'caller' '.caller')
  rc=$?
  check "an empty indexer band is no failure" '[ "$rc" = 0 ] && [ -z "$out" ]'
  exit "$failed"
) || failed=1

# A chain younger than the rate probe counts from its first block.
(
  load
  : >"$work/bands"
  graphql() {
    case $1 in
    *latestBlockHeight*) echo '{"data":{"latestBlockHeight":500}}' ;;
    *getBlocks*height*eq:\ -*) echo '{"data":{"getBlocks":null}}' ;;
    *getBlocks*) echo '{"data":{"getBlocks":[{"time":"2026-11-03T00:00:00Z"}]}}' ;;
    *getTransactions*)
      grep -oE 'gt: -?[0-9]+, lt: [0-9]+' <<<"$1" >>"$work/bands"
      echo '{"data":{"getTransactions":[]}}'
      ;;
    esac
  }
  scan_week 'MsgCall: {}' 'caller' '.caller' >/dev/null
  rc=$?
  check "a chain younger than the probe is scanned" '[ "$rc" = 0 ]'
  check "from its first block" 'grep -qx "gt: 0, lt: 501" "$work/bands"'
  exit "$failed"
) || failed=1

# A newer deploy that is parked or refused does not hide an older live one.
(
  load
  scan_week() { printf '30\tgno.land/r/alice/app/v2\n20\tgno.land/r/alice/app\n10\tgno.land/r/bob/x\n'; }
  abci() { [ "$3" != gno.land/r/alice/app/v2 ]; }
  block_time() { echo 1762128000; }
  out=$(source_deployed)
  check "an older live deploy of a project shows" 'awk -F"\t" '"'"'$2 == "alice/app" && $3 == "/r/alice/app" { f = 1 } END { exit !f }'"'"' <<<"$out"'
  check "one row per project" '[ "$(grep -c alice/app <<<"$out")" = 1 ]'
  check "the refused deploy does not" '! grep -q "app/v2" <<<"$out"'
  check "nor hides another project" 'grep -q "/r/bob/x" <<<"$out"'
  exit "$failed"
) || failed=1

# A project that deployed many packages does not crowd the others out.
(
  load
  scan_week() {
    for i in $(seq 40 -1 16); do printf '%s\tgno.land/r/busy/app/p%s\n' "$i" "$i"; done
    for p in a b c d e; do printf '10\tgno.land/r/%s/app\n' "$p"; done
  }
  abci() { echo ok; }
  block_time() { echo 1762128000; }
  out=$(source_deployed)
  check "LIST_SIZE projects despite a busy one" '[ "$(grep -c . <<<"$out")" = "$LIST_SIZE" ]'
  check "the busy one once" '[ "$(grep -c "/r/busy/" <<<"$out")" = 1 ]'
  exit "$failed"
) || failed=1

# Marker pairs must neither nest nor interleave.
(
  load
  page() { printf '%s\n' "$@" >"$work/page.md"; }
  page "$START" "$END" "$TRENDING_START" "$TRENDING_END"
  check "markers in order pass" 'page_ok "$work/page.md"'
  page "$TRENDING_START" "$TRENDING_END" "$START" "$END"
  check "trending first passes" 'page_ok "$work/page.md"'
  page "$START" "$TRENDING_START" "$TRENDING_END" "$END"
  check "trending inside whats-new fails" '! page_ok "$work/page.md"'
  page "$TRENDING_START" "$START" "$END" "$TRENDING_END"
  check "whats-new inside trending fails" '! page_ok "$work/page.md"'
  page "$START" "$TRENDING_START" "$END" "$TRENDING_END"
  check "interleaved pairs fail" '! page_ok "$work/page.md"'
  exit "$failed"
) || failed=1

# A mistyped flag is refused, not read as the page.
(
  printf '%s\n%s\n' '<!-- whats-new:start -->' '<!-- whats-new:end -->' >"$work/home.md"
  cp "$work/home.md" "$work/before.md"
  mkdir -p "$work/bin"
  printf '#!/bin/sh\nexit 22\n' >"$work/bin/curl"
  chmod +x "$work/bin/curl"
  err=$(PATH="$work/bin:$PATH" STATE_DIR="$work/state-flag" INDEXER_URL= ON_CHANGE='touch '"$work"'/changed' \
    bash "$script" --dryrn "$work/home.md" 2>&1 >/dev/null)
  code=$?
  check "a mistyped flag exits 2" '[ "$code" = 2 ]'
  check "and says so" 'grep -q "unknown option --dryrn" <<<"$err"'
  check "and leaves the page and ON_CHANGE alone" 'cmp -s "$work/home.md" "$work/before.md" && [ ! -e "$work/changed" ]'
  exit "$failed"
) || failed=1

# One break after the What's new block, none between its rows.
(
  load
  printf '20261101\thttps://x.com/_gnoland/status/1\tA post on X\n' >"$STATE_DIR/x"
  printf '20261101\t/r/gnoland/blog:p/a\tA blog post\n' >"$STATE_DIR/blog"
  out=$(render_block)
  check "one rule in the block" '[ "$(grep -cx -- "---" <<<"$out")" = 1 ]'
  check "after its last row" '[ "$(grep -v "^$" <<<"$out" | tail -n 1)" = "---" ]'
  exit "$failed"
) || failed=1

# A deployed list a week old ages out, as the ranking does.
(
  load
  printf '20261001\talice/app\t/r/alice/app\n' >"$STATE_DIR/deployed"
  check "a fresh list shows" '[ -n "$(render_deployed)" ]'
  touch -t 202001010000 "$STATE_DIR/deployed"
  check "a week-old list does not" '[ -z "$(render_deployed)" ]'
  exit "$failed"
) || failed=1

# A read-only page stays read-only, a writable one writable. Only root tells
# the mode bit from -w, so the read-only half runs as root only.
(
  printf '%s\n%s\n' '<!-- whats-new:start -->' '<!-- whats-new:end -->' >"$work/ro.md"
  mkdir -p "$work/ro-state"
  printf '20261101\thttps://x.com/_gnoland/status/1\tA post on X\n' >"$work/ro-state/x"
  touch "$work/ro-state/x.tried" "$work/ro-state/blog.tried" "$work/ro-state/release.tried" \
    "$work/ro-state/imported.tried" "$work/ro-state/shipped.tried"
  mode() { find "$1" -perm -200 | grep -q . && echo writable || echo read-only; }
  run() { STATE_DIR="$work/ro-state" INDEXER_URL= ON_CHANGE= bash "$script" "$1" >/dev/null 2>&1; }
  cp "$work/ro.md" "$work/rw.md"
  run "$work/rw.md"
  check "a writable page stays writable" '[ "$(mode "$work/rw.md")" = writable ]'
  if [ "$(id -u)" = 0 ]; then
    chmod 0444 "$work/ro.md"
    run "$work/ro.md"
    check "a read-only page stays read-only under root" '[ "$(mode "$work/ro.md")" = read-only ]'
  else
    printf 'skip a read-only page stays read-only under root (not root)\n'
  fi
  exit "$failed"
) || failed=1

# The version line names the binary the chain runs, a release tag only.
(
  load
  status() { printf '{"result":{"build_version":"%s","node_info":{"version":"v1.0.0-rc.0"}}}' "$1"; }
  fake_curl() { status "$VERSION"; }
  CURL=(fake_curl)
  check "the running candidate" '[ "$(VERSION=v1.7.0-rc.1 source_version)" = v1.7.0-rc.1 ]'
  check "a final release" '[ "$(VERSION=v1.7.0 source_version)" = v1.7.0 ]'
  check "no other string" '! VERSION=v1.7.0-dirty source_version >/dev/null'
  exit "$failed"
) || failed=1

# The page's deploy block makes a realm to deploy.
(
  mkdir -p "$work/deploy/bin"
  printf '#!/bin/sh\nexit 0\n' >"$work/deploy/bin/gnokey"
  printf '#!/bin/sh\n[ "$1 $2" = "mod init" ] && printf "module = \\"%%s\\"\\n" "$3" >gnomod.toml\n' >"$work/deploy/bin/gno"
  chmod +x "$work/deploy/bin/gnokey" "$work/deploy/bin/gno"
  awk '/^## .*Deploy from your terminal/ { on = 1 } on && /^```sh$/ { inside = 1; next } inside && /^```$/ { exit } inside' \
    "$here/home.testnet.md" >"$work/deploy/block.sh"
  (cd "$work/deploy" && PATH="$work/deploy/bin:$PATH" bash -e block.sh >/dev/null 2>&1)
  rc=$?
  check "the deploy block runs, with gnokey and gno stubbed" '[ "$rc" = 0 ]'
  check "and writes a realm" 'grep -q "^package myrealm$" "$work/deploy/myrealm/myrealm.gno"'
  check "with its module path" 'grep -q "gno.land/r/<your-g1-addr>/myrealm" "$work/deploy/myrealm/gnomod.toml"'
  exit "$failed"
) || failed=1

[ "$failed" = 0 ] && echo PASS || {
  echo FAIL
  exit 1
}
