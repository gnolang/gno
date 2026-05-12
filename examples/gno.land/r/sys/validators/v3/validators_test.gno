// Tests for stateless v3.
//
// State isolation: gno's test runner constructs a fresh testParams
// per top-level TestXxx, but t.Run subtests share the same map.
// Each subtest below seeds (or clears) the valset slots it cares
// about up front; do the same when adding cases.
package validators

import (
	"chain"
	"testing"

	"gno.land/p/nt/urequire/v0"
)

const (
	// Three distinct deterministic test pubkeys (and the addresses
	// they derive to) used across the cases below. Stable across
	// runs.
	pubKeyA = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq3ds6sdvc0shfkq02h6xx5g0jp04aadexfnpsmgjxu72xz9y30aqfrlpny"
	pubKeyB = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
	pubKeyC = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqddddqg2glc8x4fl7vxjlnr7p5a3czm5kcdp4239sg6yqdc4rc2r5cjrffs"

	module    = "node"
	submodule = "valset"
	currKey   = "current"
	propKey   = "proposed"
	dirtyKey  = "dirty"
)

func resetValset(t *testing.T) {
	t.Helper()
	testing.SetSysParamStrings(module, submodule, currKey, []string{})
	testing.SetSysParamStrings(module, submodule, propKey, []string{})
	testing.SetSysParamBool(module, submodule, dirtyKey, false)
}

func mustAddr(t *testing.T, pk string) address {
	t.Helper()
	a, err := chain.PubKeyAddress(pk)
	urequire.NoError(t, err, "valid pubkey")
	return a
}

func TestEffectiveView_DirtyFalseReadsCurrent(t *testing.T) {
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	addrA := mustAddr(t, pubKeyA)
	urequire.True(t, IsValidator(addrA), "A in current")
	urequire.Equal(t, 1, len(GetValidators()))
}

func TestEffectiveView_DirtyTrueReadsProposed(t *testing.T) {
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})
	testing.SetSysParamStrings(module, submodule, propKey, []string{pubKeyA + ":10", pubKeyB + ":1"})
	testing.SetSysParamBool(module, submodule, dirtyKey, true)

	addrA := mustAddr(t, pubKeyA)
	addrB := mustAddr(t, pubKeyB)
	urequire.True(t, IsValidator(addrA), "A still visible")
	urequire.True(t, IsValidator(addrB), "B visible via proposed")
	urequire.Equal(t, 2, len(GetValidators()))
}

func TestGetValidator_ReturnsParsedEntry(t *testing.T) {
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":42"})

	v := GetValidator(mustAddr(t, pubKeyA))
	urequire.Equal(t, uint64(42), v.VotingPower)
	urequire.Equal(t, pubKeyA, v.PubKey)
}

func TestRender_DefaultTestHeight(t *testing.T) {
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	out := Render("")
	// gnovm/pkg/test/test.go sets DefaultHeight = 123.
	urequire.Equal(t,
		"## Valset at height 123\n\n- #0: "+mustAddr(t, pubKeyA).String()+" (10)\n",
		out)
}

// The same-block accumulation, panic-on-remove-of-missing,
// panic-on-add-of-existing, and pubkey-validation tests for the
// signing-keyed NewProposalRequest/newValsetChangeExecutor were
// removed alongside the function itself. Equivalent coverage for
// the operator-keyed flow lives in proposal_test.gno
// (TestNewValidatorProposalRequest_*) and cache_test.gno
// (TestRotateValoperSigningKey_AccumulatesAcrossSameBlock).
//
// Empty-final-set rejection in the executor still applies; it's
// covered by TestNewValidatorProposalRequest_RemoveOperator in
// proposal_test.gno.
