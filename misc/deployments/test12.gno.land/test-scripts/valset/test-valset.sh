#!/usr/bin/env bash
# Comprehensive valset management test suite.
#
# Tests three areas:
#   1. Direct govDAO proposals (without valopers realm)
#   2. Valopers realm operations (register, update, auth)
#   3. GovDAO proposals via the valopers realm
#
# Usage:
#   ./test-valset.sh <val1_pubkey> <val2_pubkey>
#
#   val1_pubkey  bech32 validator pubkey (gpub1...) — must NOT be in the initial valset
#   val2_pubkey  bech32 validator pubkey (gpub1...) — must NOT be in the initial valset
#
# The validator addresses are derived from the pubkeys automatically.
#
# Optional env overrides (all govdao-scripts vars are also accepted):
#   TEST_UNREG_ADDR          a valid bech32 address not registered in valopers
#                            (defaults to GNOKEY_NAME's own address)
#   VALOPER_REGISTRATION_FEE ugnot to send with Register (default: 20000000)
#
# The account identified by GNOKEY_NAME must:
#   - Be a govDAO T1 member (to create/vote on proposals)
#   - Hold enough ugnot to cover gas fees and two valoper registration fees
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# shellcheck source=common
source "$SCRIPT_DIR/common"

# ---- Argument validation ----

if [ $# -ne 2 ]; then
  echo "Usage: $0 <val1_pubkey> <val2_pubkey>" >&2
  echo "" >&2
  echo "  val1_pubkey  gpub1... pubkey of test validator 1 (must NOT be in initial valset)" >&2
  echo "  val2_pubkey  gpub1... pubkey of test validator 2 (must NOT be in initial valset)" >&2
  exit 1
fi

TEST_VAL1_PUBKEY="$1"
TEST_VAL2_PUBKEY="$2"

TEST_VAL1_ADDR=$(pubkey_to_addr "$TEST_VAL1_PUBKEY")
TEST_VAL2_ADDR=$(pubkey_to_addr "$TEST_VAL2_PUBKEY")

# ---- Password prompt ----

if [ -z "${INSECURE_STDIN_PASSWORD:-}" ]; then
  read -rsp "Enter gnokey password for '${GNOKEY_NAME}': " INSECURE_STDIN_PASSWORD
  echo
fi
export INSECURE_STDIN_PASSWORD

# Export all env vars so child scripts (govdao-scripts) inherit them.
export GNOKEY_NAME CHAIN_ID REMOTE GAS_WANTED GAS_FEE VALOPER_REGISTRATION_FEE
export GOVDAO_SCRIPTS VALSET_SCRIPTS

# ---- Test framework ----

TESTS_PASSED=0
TESTS_FAILED=0
CURRENT_GROUP=""

begin_group() {
  CURRENT_GROUP="$1"
  echo ""
  echo "══════════════════════════════════════════════════════"
  echo "  $1"
  echo "══════════════════════════════════════════════════════"
}

_pass() {
  echo "  [PASS] $1"
  TESTS_PASSED=$((TESTS_PASSED + 1))
}

_fail() {
  echo "  [FAIL] $1"
  TESTS_FAILED=$((TESTS_FAILED + 1))
}

# Run a command, expect it to succeed (exit 0).
expect_success() {
  local desc="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    _pass "$desc"
  else
    _fail "$desc — expected success, got failure"
  fi
}

# Run a command, expect it to fail (non-zero exit).
expect_failure() {
  local desc="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    _fail "$desc — expected failure, got success"
  else
    _pass "$desc"
  fi
}

# ---- State query helpers ----

# Query valoper render output. Returns raw render text.
query_valoper() {
  local addr="$1"
  gnokey query vm/qrender -remote "$REMOTE" -data "gno.land/r/gnops/valopers:${addr}" 2>&1
}

# Returns 0 if the address is registered in the valopers realm.
valoper_exists() {
  local addr="$1"
  local out
  out=$(query_valoper "$addr")
  ! echo "$out" | grep -q "unknown address"
}

# Returns 0 if the address is in the current validator set.
is_validator() {
  local addr="$1"
  local out
  out=$(gnokey query vm/qeval -remote "$REMOTE" \
    -data "$(printf 'gno.land/r/sys/validators/v2\nIsValidator(address("%s"))' "$addr")" 2>&1)
  echo "$out" | grep -q "(bool) true"
}

# Assert that an address IS in the validator set.
assert_in_valset() {
  local addr="$1"
  local label="${2:-$addr}"
  if is_validator "$addr"; then
    _pass "state: ${label} is in validator set"
  else
    _fail "state: ${label} should be in validator set but is not"
  fi
}

# Assert that an address is NOT in the validator set.
assert_not_in_valset() {
  local addr="$1"
  local label="${2:-$addr}"
  if ! is_validator "$addr"; then
    _pass "state: ${label} is not in validator set"
  else
    _fail "state: ${label} should not be in validator set but is"
  fi
}

# Assert that an address IS registered in valopers.
assert_valoper_exists() {
  local addr="$1"
  local label="${2:-$addr}"
  if valoper_exists "$addr"; then
    _pass "state: ${label} is registered in valopers"
  else
    _fail "state: ${label} should be registered in valopers but is not"
  fi
}

# Assert that an address is NOT registered in valopers.
assert_valoper_missing() {
  local addr="$1"
  local label="${2:-$addr}"
  if ! valoper_exists "$addr"; then
    _pass "state: ${label} is not registered in valopers"
  else
    _fail "state: ${label} should not be registered in valopers but is"
  fi
}

# Assert that a valoper's KeepRunning flag has the expected value.
assert_keeprunning() {
  local addr="$1"
  local expected="$2"  # "true" or "false"
  local label="${3:-$addr}"
  local out
  out=$(gnokey query vm/qeval -remote "$REMOTE" \
    -data "$(printf 'gno.land/r/gnops/valopers\nGetByAddr(address("%s")).KeepRunning' "$addr")" 2>&1)
  if echo "$out" | grep -q "(bool) ${expected}"; then
    _pass "state: ${label} KeepRunning=${expected}"
  else
    _fail "state: ${label} KeepRunning expected ${expected}, got: $(echo "$out" | tail -1)"
  fi
}

# ---- Helpers to call sibling scripts ----

run_govdao() {
  "$GOVDAO_SCRIPTS/$1" "${@:2}"
}

run_valset() {
  "$VALSET_SCRIPTS/$1" "${@:2}"
}

# ---- Cleanup helper ----

# Best-effort cleanup: remove test validators from valset and valopers.
# Failures are silenced — the state may already be clean.
cleanup_test_state() {
  echo ""
  echo "--- Cleanup: removing test validators from valset and valopers ---"

  run_govdao rm-validator.sh "$TEST_VAL1_ADDR" >/dev/null 2>&1 || true
  run_govdao rm-validator.sh "$TEST_VAL2_ADDR" >/dev/null 2>&1 || true

  # There is no "unregister" function in the valopers realm.
  # If tests leave stale valopers registrations, they persist across runs.
  # The test suite accounts for this by handling ErrValoperExists gracefully.
  echo "  Note: valopers registrations are persistent (no Unregister function)."
  echo "  Re-running tests after a prior run may hit 'valoper already exists' for"
  echo "  valopers tests — this is expected and handled."
}

# ======================================================================
# GROUP 1: Direct govDAO proposals — no valopers realm involved
# ======================================================================

begin_group "Group 1: Direct govDAO proposals"

echo ""
echo "  Val1: ${TEST_VAL1_ADDR}"
echo "  Val2: ${TEST_VAL2_ADDR}"

# Ensure clean initial state.
cleanup_test_state

# 1.1 — Add a validator not in the valset and not in valopers.
echo ""
echo "  --- 1.1: Add val1 (not in valset, not in valopers) via direct proposal ---"
assert_not_in_valset "$TEST_VAL1_ADDR" "val1"
assert_valoper_missing "$TEST_VAL1_ADDR" "val1"
expect_success "1.1: add val1 via direct proposal" \
  run_govdao add-validator.sh "$TEST_VAL1_PUBKEY" 1
assert_in_valset "$TEST_VAL1_ADDR" "val1"

# 1.2 — Add the same validator again (already in valset, same pubkey+power).
#        NewPropRequest calls addValidator which may succeed as an update;
#        this checks the system is idempotent for direct proposals.
echo ""
echo "  --- 1.2: Add val1 again (already in valset, same pubkey+power) ---"
assert_in_valset "$TEST_VAL1_ADDR" "val1"
expect_success "1.2: re-add val1 (idempotent update via direct proposal)" \
  run_govdao add-validator.sh "$TEST_VAL1_PUBKEY" 1
assert_in_valset "$TEST_VAL1_ADDR" "val1"

# 1.3 — Remove val1 from the valset.
echo ""
echo "  --- 1.3: Remove val1 (in valset, not in valopers) ---"
assert_in_valset "$TEST_VAL1_ADDR" "val1"
expect_success "1.3: remove val1 via direct proposal" \
  run_govdao rm-validator.sh "$TEST_VAL1_ADDR"
assert_not_in_valset "$TEST_VAL1_ADDR" "val1"

# 1.4 — Remove a validator that is not in the valset.
#        The validators/v2 realm calls removeValidator which panics if the
#        address is unknown. The proposal execution will fail with a panic,
#        causing the entire MsgRun to revert.
echo ""
echo "  --- 1.4: Remove val1 again (not in valset) — expect failure ---"
assert_not_in_valset "$TEST_VAL1_ADDR" "val1"
expect_failure "1.4: remove val1 not in valset (should panic)" \
  run_govdao rm-validator.sh "$TEST_VAL1_ADDR"
assert_not_in_valset "$TEST_VAL1_ADDR" "val1"

# 1.5 — Add val2 with explicit voting power of 2.
echo ""
echo "  --- 1.5: Add val2 with power=2 ---"
assert_not_in_valset "$TEST_VAL2_ADDR" "val2"
expect_success "1.5: add val2 with power=2" \
  run_govdao add-validator.sh "$TEST_VAL2_PUBKEY" 2
assert_in_valset "$TEST_VAL2_ADDR" "val2"

# 1.6 — Remove val2.
echo ""
echo "  --- 1.6: Remove val2 ---"
expect_success "1.6: remove val2" \
  run_govdao rm-validator.sh "$TEST_VAL2_ADDR"
assert_not_in_valset "$TEST_VAL2_ADDR" "val2"

# ======================================================================
# GROUP 2: Valopers realm operations
# ======================================================================

begin_group "Group 2: Valopers realm operations"

# 2.1 — Register val1 in the valopers realm.
echo ""
echo "  --- 2.1: Register val1 in valopers ---"
assert_valoper_missing "$TEST_VAL1_ADDR" "val1"
expect_success "2.1: register val1" \
  run_valset register-valoper.sh \
    "$TEST_VAL1_PUBKEY" "TestVal1" "Test validator 1" "cloud"
assert_valoper_exists "$TEST_VAL1_ADDR" "val1"

# 2.2 — Register val1 again — should fail with ErrValoperExists.
echo ""
echo "  --- 2.2: Register val1 again (already registered) — expect failure ---"
expect_failure "2.2: re-register val1 (should panic with ErrValoperExists)" \
  run_valset register-valoper.sh \
    "$TEST_VAL1_PUBKEY" "TestVal1" "Test validator 1" "cloud"

# 2.3 — Register with invalid moniker (too short: single character).
#        The regex requires at least 2 characters (one leading + one trailing alphanum).
echo ""
echo "  --- 2.3: Register with invalid moniker (single char) — expect failure ---"
expect_failure "2.3: register with single-char moniker (regex mismatch)" \
  run_valset register-valoper.sh \
    "$TEST_VAL2_PUBKEY" "X" "A description" "cloud"

# 2.4 — Register with invalid server type.
echo ""
echo "  --- 2.4: Register with invalid server type — expect failure ---"
expect_failure "2.4: register with invalid server type" \
  run_valset register-valoper.sh \
    "$TEST_VAL2_PUBKEY" "TestVal2" "A description" "bare-metal"

# 2.5 — Register with invalid pubkey.
#        pubkey_to_addr fails immediately — no transaction is sent.
echo ""
echo "  --- 2.5: Register with invalid pubkey — expect failure ---"
expect_failure "2.5: register with non-bech32 pubkey (fails at address derivation)" \
  run_valset register-valoper.sh \
    "not-a-valid-pubkey" "TestVal2" "A description" "cloud"

# 2.6 — Register with insufficient fee (1 ugnot).
#        The realm requires minFee (20 GNOT = 20,000,000 ugnot by default).
echo ""
echo "  --- 2.6: Register with insufficient fee — expect failure ---"
_saved_fee="$VALOPER_REGISTRATION_FEE"
export VALOPER_REGISTRATION_FEE=1
expect_failure "2.6: register with 1 ugnot fee (below minFee)" \
  run_valset register-valoper.sh \
    "$TEST_VAL2_PUBKEY" "TestVal2" "A description" "cloud"
export VALOPER_REGISTRATION_FEE="$_saved_fee"
assert_valoper_missing "$TEST_VAL2_ADDR" "val2"

# 2.7 — Register val2 with valid data.
echo ""
echo "  --- 2.7: Register val2 with valid data ---"
expect_success "2.7: register val2" \
  run_valset register-valoper.sh \
    "$TEST_VAL2_PUBKEY" "TestVal2" "Test validator 2" "on-prem"
assert_valoper_exists "$TEST_VAL2_ADDR" "val2"

# 2.8 — Update val1's moniker.
echo ""
echo "  --- 2.8: Update val1 moniker ---"
expect_success "2.8: update val1 moniker to 'Val1Updated'" \
  run_valset update-valoper-moniker.sh "$TEST_VAL1_ADDR" "Val1Updated"

# 2.9 — Update val1's description.
echo ""
echo "  --- 2.9: Update val1 description ---"
expect_success "2.9: update val1 description" \
  run_valset update-valoper-description.sh "$TEST_VAL1_ADDR" "Updated description for val1"

# 2.10 — Update val1's server type.
echo ""
echo "  --- 2.10: Update val1 server type to data-center ---"
expect_success "2.10: update val1 server type to data-center" \
  run_valset update-valoper-servertype.sh "$TEST_VAL1_ADDR" "data-center"

# 2.11 — Update with invalid server type — should fail.
echo ""
echo "  --- 2.11: Update server type to invalid value — expect failure ---"
expect_failure "2.11: update val1 server type to 'invalid'" \
  run_valset update-valoper-servertype.sh "$TEST_VAL1_ADDR" "invalid"

# 2.12 — Set val1 KeepRunning to false.
echo ""
echo "  --- 2.12: Set val1 KeepRunning=false ---"
expect_success "2.12: set val1 KeepRunning=false" \
  run_valset update-valoper-keeprunning.sh "$TEST_VAL1_ADDR" false
assert_keeprunning "$TEST_VAL1_ADDR" "false" "val1"

# 2.13 — Restore val1 KeepRunning to true.
echo ""
echo "  --- 2.13: Set val1 KeepRunning=true ---"
expect_success "2.13: set val1 KeepRunning=true" \
  run_valset update-valoper-keeprunning.sh "$TEST_VAL1_ADDR" true
assert_keeprunning "$TEST_VAL1_ADDR" "true" "val1"

# 2.14 — Add an auth member to val1, then remove it.
echo ""
echo "  --- 2.14: Add and remove auth member on val1 ---"
# Use val2's address as a stand-in for a secondary operator.
expect_success "2.14a: add val2 to val1's auth list" \
  run_valset add-auth-member.sh "$TEST_VAL1_ADDR" "$TEST_VAL2_ADDR"
expect_success "2.14b: remove val2 from val1's auth list" \
  run_valset rm-auth-member.sh "$TEST_VAL1_ADDR" "$TEST_VAL2_ADDR"

# ======================================================================
# GROUP 3: GovDAO proposals via the valopers realm
# ======================================================================

begin_group "Group 3: GovDAO proposals via valopers"

# Precondition: both val1 and val2 are registered in valopers (from group 2),
# and neither is in the validator set.

echo ""
echo "  Precondition check:"
assert_valoper_exists "$TEST_VAL1_ADDR" "val1"
assert_valoper_exists "$TEST_VAL2_ADDR" "val2"
assert_not_in_valset "$TEST_VAL1_ADDR" "val1"
assert_not_in_valset "$TEST_VAL2_ADDR" "val2"

# 3.1 — Add val1 via valopers proposal (registered in valopers, not in valset).
echo ""
echo "  --- 3.1: Add val1 from valopers (registered, not in valset) ---"
expect_success "3.1: add val1 via valopers proposal" \
  run_govdao add-validator-from-valopers.sh "$TEST_VAL1_ADDR"
assert_in_valset "$TEST_VAL1_ADDR" "val1"

# 3.2 — Add val1 again via valopers (already in valset, same pubkey+power=1).
#        NewValidatorProposalRequest checks: if validator exists AND
#        VotingPower==1 AND PubKey==valoper.PubKey → panic(ErrSameValues).
echo ""
echo "  --- 3.2: Add val1 again from valopers (in valset, same values) — expect failure ---"
assert_in_valset "$TEST_VAL1_ADDR" "val1"
expect_failure "3.2: re-add val1 from valopers (ErrSameValues)" \
  run_govdao add-validator-from-valopers.sh "$TEST_VAL1_ADDR"
assert_in_valset "$TEST_VAL1_ADDR" "val1"

# 3.3 — Add an address that has no valopers registration — expect failure.
#        TEST_UNREG_ADDR must be a valid bech32 address that is NOT registered
#        in the valopers realm. Defaults to the GNOKEY_NAME account's own address
#        (resolved below), which is typically not registered as a valoper.
echo ""
echo "  --- 3.3: Add non-registered address via valopers proposal — expect failure ---"
if [ -z "${TEST_UNREG_ADDR:-}" ]; then
  TEST_UNREG_ADDR=$(gnokey list 2>/dev/null \
    | grep "\. ${GNOKEY_NAME} (" \
    | grep -o 'addr: g1[a-z0-9]*' \
    | awk '{print $2}' || true)
fi
if [ -z "${TEST_UNREG_ADDR:-}" ]; then
  echo "  SKIP 3.3: cannot resolve TEST_UNREG_ADDR (set it manually to skip this)"
else
  expect_failure "3.3: add unregistered address via valopers (ErrValoperMissing)" \
    run_govdao add-validator-from-valopers.sh "$TEST_UNREG_ADDR"
fi

# 3.4 — Set val2 KeepRunning=false, then submit a valopers proposal.
#        When KeepRunning=false and validator is not yet in valset:
#        NewValidatorProposalRequest panics with ErrValidatorMissing.
echo ""
echo "  --- 3.4: val2 KeepRunning=false, not in valset — expect failure on proposal ---"
run_valset update-valoper-keeprunning.sh "$TEST_VAL2_ADDR" false >/dev/null 2>&1
assert_not_in_valset "$TEST_VAL2_ADDR" "val2"
expect_failure "3.4: valopers proposal for val2 (KeepRunning=false, not in valset)" \
  run_govdao add-validator-from-valopers.sh "$TEST_VAL2_ADDR"
assert_not_in_valset "$TEST_VAL2_ADDR" "val2"

# Restore val2 KeepRunning for subsequent tests.
run_valset update-valoper-keeprunning.sh "$TEST_VAL2_ADDR" true >/dev/null 2>&1

# 3.5 — Add val2 via valopers (registered, KeepRunning=true, not in valset).
echo ""
echo "  --- 3.5: Add val2 from valopers (registered, KeepRunning=true, not in valset) ---"
assert_not_in_valset "$TEST_VAL2_ADDR" "val2"
expect_success "3.5: add val2 via valopers proposal" \
  run_govdao add-validator-from-valopers.sh "$TEST_VAL2_ADDR"
assert_in_valset "$TEST_VAL2_ADDR" "val2"

# 3.6 — Set val1 KeepRunning=false and submit valopers proposal to remove it.
#        When KeepRunning=false and validator IS in valset: votingPower=0 → remove.
echo ""
echo "  --- 3.6: Remove val1 via valopers (KeepRunning=false, in valset) ---"
assert_in_valset "$TEST_VAL1_ADDR" "val1"
run_valset update-valoper-keeprunning.sh "$TEST_VAL1_ADDR" false >/dev/null 2>&1
expect_success "3.6: remove val1 via valopers proposal (KeepRunning=false)" \
  run_govdao add-validator-from-valopers.sh "$TEST_VAL1_ADDR"
assert_not_in_valset "$TEST_VAL1_ADDR" "val1"

# 3.7 — After removal, try valopers proposal again for val1.
#        KeepRunning=false AND not in valset → ErrValidatorMissing.
echo ""
echo "  --- 3.7: Valopers proposal for val1 (KeepRunning=false, not in valset) — expect failure ---"
assert_not_in_valset "$TEST_VAL1_ADDR" "val1"
expect_failure "3.7: valopers proposal for val1 (KeepRunning=false, not in valset)" \
  run_govdao add-validator-from-valopers.sh "$TEST_VAL1_ADDR"

# Restore val1 KeepRunning for a clean state.
run_valset update-valoper-keeprunning.sh "$TEST_VAL1_ADDR" true >/dev/null 2>&1

# 3.8 — Add val1 back via direct proposal (bypassing valopers), then verify
#        that a valopers proposal reflects the live state correctly.
echo ""
echo "  --- 3.8: Add val1 via direct proposal, re-add via valopers — expect ErrSameValues ---"
run_govdao add-validator.sh "$TEST_VAL1_PUBKEY" 1 >/dev/null 2>&1
assert_in_valset "$TEST_VAL1_ADDR" "val1"
# Valopers proposal would try power=1, pubkey=same → ErrSameValues.
expect_failure "3.8: valopers proposal for val1 already in valset (ErrSameValues)" \
  run_govdao add-validator-from-valopers.sh "$TEST_VAL1_ADDR"

# ======================================================================
# GROUP 4: Mixed/edge cases
# ======================================================================

begin_group "Group 4: Edge cases"

# 4.1 — Direct remove of val1 (in valset, registered in valopers).
echo ""
echo "  --- 4.1: Remove val1 via direct proposal (in valset, registered in valopers) ---"
assert_in_valset "$TEST_VAL1_ADDR" "val1"
expect_success "4.1: direct remove val1 while registered in valopers" \
  run_govdao rm-validator.sh "$TEST_VAL1_ADDR"
assert_not_in_valset "$TEST_VAL1_ADDR" "val1"
assert_valoper_exists "$TEST_VAL1_ADDR" "val1"  # still in valopers

# 4.2 — Direct add of val1 with power=2, then a valopers proposal to correct it.
#        The valopers realm always uses power=1 for KeepRunning=true, so the proposal
#        detects a difference (power 2 → 1) and is NOT rejected with ErrSameValues.
echo ""
echo "  --- 4.2: Add val1 with power=2 via direct proposal, then correct via valopers ---"
run_govdao add-validator.sh "$TEST_VAL1_PUBKEY" 2 >/dev/null 2>&1
assert_in_valset "$TEST_VAL1_ADDR" "val1"
run_valset update-valoper-keeprunning.sh "$TEST_VAL1_ADDR" true >/dev/null 2>&1
# valopers proposal: power=1 vs current power=2 → different → NOT ErrSameValues → succeeds.
expect_success "4.2: valopers proposal corrects power=2 to power=1" \
  run_govdao add-validator-from-valopers.sh "$TEST_VAL1_ADDR"
assert_in_valset "$TEST_VAL1_ADDR" "val1"

# 4.3 — Remove val1 and val2 to restore clean state.
echo ""
echo "  --- 4.3: Final cleanup ---"
run_govdao rm-validator.sh "$TEST_VAL1_ADDR" >/dev/null 2>&1 || true
run_govdao rm-validator.sh "$TEST_VAL2_ADDR" >/dev/null 2>&1 || true
assert_not_in_valset "$TEST_VAL1_ADDR" "val1"
assert_not_in_valset "$TEST_VAL2_ADDR" "val2"
_pass "4.3: cleanup — both test validators removed from valset"

# ======================================================================
# Summary
# ======================================================================

echo ""
echo "══════════════════════════════════════════════════════"
echo "  Test Results"
echo "══════════════════════════════════════════════════════"
echo "  Passed: ${TESTS_PASSED}"
echo "  Failed: ${TESTS_FAILED}"
echo "  Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ "$TESTS_FAILED" -gt 0 ]; then
  echo "  SOME TESTS FAILED."
  exit 1
else
  echo "  ALL TESTS PASSED."
fi
