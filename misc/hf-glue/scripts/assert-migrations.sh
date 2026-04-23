#!/usr/bin/env bash
# assert-migrations.sh — verify that the hf-glue post-replay state matches
# the intent of every migration step (see
# misc/deployments/gnoland-1/migrations/). Run against a live node after
# genesis replay has completed.
#
# Why: with --skip-failing-genesis-txs, any migration that silently panics is
# absorbed without aborting the chain, so a defective rc can boot and look
# fine while leaving the T1 rotation, v3 deploy, or param flips unapplied.
# This script is the positive check that catches silent migration misfires.
#
# Env
# ===
#   REMOTE                 RPC endpoint to query (default http://localhost:26657)
#   EXPECTED_T1            bech32 address expected to be the sole post-rotation
#                          T1 member. When rotation is configured, this is the
#                          value of NEW_T1_ADDR used at genesis build time.
#                          Defaults to gnoland-1's production T1 (g1aeddlft…),
#                          override when building from a different CALLER.
#   EXPECTED_VALSET_ADDRS  Space-separated list of bech32 g1 addresses that
#                          should appear in r/sys/validators/v2 after
#                          migration 01. Leave empty to skip the v2-valset
#                          check.
#   GNOKEY_BIN             gnokey binary (default: gnokey on $PATH)
#
# Exit status
# ===========
#   0  — all assertions passed
#   1  — one or more assertions failed (details printed to stdout)
#   2  — prerequisite error (bad tool, RPC unreachable, etc.)
set -euo pipefail

REMOTE="${REMOTE:-http://localhost:26657}"
EXPECTED_T1="${EXPECTED_T1:-g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m}"
EXPECTED_VALSET_ADDRS="${EXPECTED_VALSET_ADDRS:-}"
GNOKEY_BIN="${GNOKEY_BIN:-gnokey}"

# Manfred is the pre-rotation sole T1 on gnoland1; compared against
# $EXPECTED_T1 to decide whether to assert manfred was withdrawn.
readonly MANFRED="g1manfred47kzduec920z88wfr64ylksmdcedlf5"

command -v "$GNOKEY_BIN" >/dev/null 2>&1 || {
  echo "gnokey not found on PATH (set GNOKEY_BIN=...)" >&2
  exit 2
}

fail=0
pass=0

# ---- Helpers

# Extracts the `data:` line from a `gnokey query` response, stripping the
# leading `data: ` prefix. Single-line extraction — if the response body spans
# multiple lines, only the first is returned.
extract_data() {
  awk 'NR==FNR{next}/^data:/{sub(/^data: /,""); print; exit}' /dev/null "$@"
}

query_param() {
  local key="$1"
  "$GNOKEY_BIN" query -remote "$REMOTE" "params/$key" 2>/dev/null |
    awk '/^data:/{sub(/^data: /,""); print; exit}'
}

query_qeval() {
  local expr="$1"
  "$GNOKEY_BIN" query -remote "$REMOTE" "vm/qeval" --data "$expr" 2>/dev/null |
    awk '/^data:/{sub(/^data: /,""); print; exit}'
}

query_qfuncs_raw() {
  local pkg="$1"
  "$GNOKEY_BIN" query -remote "$REMOTE" "vm/qfuncs" --data "$pkg" 2>/dev/null
}

check_eq() {
  local desc="$1" expected="$2" actual="$3"
  if [[ "$actual" == "$expected" ]]; then
    printf '  [OK]   %s\n' "$desc"
    pass=$((pass + 1))
  else
    printf '  [FAIL] %s\n         want=%s\n         got =%s\n' "$desc" "$expected" "$actual"
    fail=$((fail + 1))
  fi
}

check_contains() {
  local desc="$1" needle="$2" haystack="$3"
  if [[ "$haystack" == *"$needle"* ]]; then
    printf '  [OK]   %s\n' "$desc"
    pass=$((pass + 1))
  else
    printf '  [FAIL] %s\n         want substring=%s\n         in=%s\n' "$desc" "$needle" "$haystack"
    fail=$((fail + 1))
  fi
}

check_not_contains() {
  local desc="$1" needle="$2" haystack="$3"
  if [[ "$haystack" != *"$needle"* ]]; then
    printf '  [OK]   %s\n' "$desc"
    pass=$((pass + 1))
  else
    printf '  [FAIL] %s\n         did NOT want substring=%s\n         in=%s\n' "$desc" "$needle" "$haystack"
    fail=$((fail + 1))
  fi
}

echo "━━━ assert-migrations against $REMOTE ━━━"
echo "  expected T1: $EXPECTED_T1"
[[ -n "$EXPECTED_VALSET_ADDRS" ]] && echo "  expected v2: $EXPECTED_VALSET_ADDRS"
echo

# ---- Sanity: node reachable
if ! "$GNOKEY_BIN" query -remote "$REMOTE" ".app/version" >/dev/null 2>&1; then
  echo "  RPC unreachable at $REMOTE" >&2
  exit 2
fi

# ---- Migration 05 + 07 (sysnames namespace check disable/restore)
# Step 05 sets the vm param to "", step 07 restores it. If 06 (addpkg v3)
# fails silently before 07 runs, the path stays empty and the authz check
# is permanently off — exactly the silent-failure case this guards against.
check_eq 'migration 05+07: vm:p:sysnames_pkgpath restored to r/sys/names' \
  '"gno.land/r/sys/names"' \
  "$(query_param 'vm:p:sysnames_pkgpath')"

# r/sys/names internal flag stayed true through the restore (we never touched
# .enabled, only the VM param pointing at the pkg).
check_eq 'r/sys/names.IsEnabled() still true after migration' \
  '(true bool)' \
  "$(query_qeval 'gno.land/r/sys/names.IsEnabled()')"

# ---- Migration 06 (addpkg r/sys/validators/v3)
v3funcs="$(query_qfuncs_raw 'gno.land/r/sys/validators/v3')"
check_contains 'migration 06: v3 realm deployed (NewValsetChangeExecutor exported)' \
  'NewValsetChangeExecutor' "$v3funcs"
check_contains 'migration 06: v3 realm deployed (GetValidators exported)' \
  'GetValidators' "$v3funcs"

# ---- Migration 08 (valset_realm_path points at v3)
check_eq 'migration 08: vm:p:valset_realm_path = v3' \
  '"gno.land/r/sys/validators/v3"' \
  "$(query_param 'vm:p:valset_realm_path')"

# ---- v3 pending-update drain
# If new_updates_available=true, EndBlocker hasn't consumed the last proposal
# yet — either the chain isn't producing blocks or the param path is wrong.
# valset_prev reflects the last applied valset; no strict assertion because
# it legitimately evolves as add-validator proposals land.
check_eq 'v3: no pending valset update unconsumed by EndBlocker' \
  'false' \
  "$(query_param 'vm:gno.land/r/sys/validators/v3:new_updates_available')"

# ---- Migration 01 (v2 valset swapped) — informational only
# v2 is vestigial after PR #5485: EndBlocker no longer reads from it. The
# reset_valset migration is cosmetic and is known to partial-apply when a
# removed-validator exists in the batch (the whole proposal panics on the
# first missing entry and leaves subsequent removes + the new-validator add
# unapplied). Surface the current state for visual inspection but do not
# fail the check — consensus is driven by GenesisDoc.Validators and v3.
if [[ -n "$EXPECTED_VALSET_ADDRS" ]]; then
  v2out="$(query_qeval 'gno.land/r/sys/validators/v2.GetValidators()')"
  for addr in $EXPECTED_VALSET_ADDRS; do
    if [[ "$v2out" == *"$addr"* ]]; then
      printf '  [OK]   migration 01 (informational): r/sys/validators/v2 has %s\n' "$addr"
      pass=$((pass + 1))
    else
      printf '  [WARN] migration 01 (informational): r/sys/validators/v2 missing %s (v2 is dead code post-PR#5485)\n' "$addr"
    fi
  done
fi

# ---- Migration 02-04 (T1 rotation)
# memberstore.Get() has an ACL that rejects calls from qeval's empty-realm
# context, so membership is read from the public render endpoints:
#   /r/gov/dao/v3/memberstore         — tier summary ("Tier T1 contains N members")
#   /r/gov/dao/v3/memberstore:members — tabular member list with T1/T2/T3 rows
summary="$("$GNOKEY_BIN" query -remote "$REMOTE" vm/qrender --data 'gno.land/r/gov/dao/v3/memberstore:' 2>/dev/null)"
members="$("$GNOKEY_BIN" query -remote "$REMOTE" vm/qrender --data 'gno.land/r/gov/dao/v3/memberstore:members' 2>/dev/null)"

t1_count="$(printf '%s\n' "$summary" | grep -oE 'Tier T1 contains [0-9]+ members' | grep -oE '[0-9]+' | head -1)"
check_eq 'migration 02-04: govDAO T1 tier size = 1' '1' "${t1_count:-<none>}"

check_contains "migration 02-04: $EXPECTED_T1 is T1 member" \
  "| $EXPECTED_T1 |" "$members"

if [[ "$EXPECTED_T1" != "$MANFRED" ]]; then
  check_not_contains 'migration 03-04: manfred withdrawn from T1' \
    "| $MANFRED |" "$members"
fi

echo
printf 'Results: %d ok, %d fail\n' "$pass" "$fail"
exit "$fail"
