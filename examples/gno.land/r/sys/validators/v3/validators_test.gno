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
	pVals "gno.land/p/sys/validators"
)

const (
	// Two distinct deterministic test pubkeys (and the addresses
	// they derive to) used across the cases below. Stable across
	// runs.
	pubKeyA = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq3ds6sdvc0shfkq02h6xx5g0jp04aadexfnpsmgjxu72xz9y30aqfrlpny"
	pubKeyB = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"

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

func TestCallback_AccumulatesOnDirtyProposed(t *testing.T) {
	// Same-block accumulation: a prior callback wrote
	// valset:proposed + dirty=true. A second callback in the same
	// block must read effective (= proposed) and apply on top of it.
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})
	// Simulate "proposal A added validator B at power 1, EndBlocker
	// has not yet promoted." dirty=true points reads at proposed.
	testing.SetSysParamStrings(module, submodule, propKey, []string{pubKeyA + ":10", pubKeyB + ":1"})
	testing.SetSysParamBool(module, submodule, dirtyKey, true)

	// Second callback removes B. Construct an executor and run its
	// callback directly (full govdao flow exercised in txtar).
	addrB := mustAddr(t, pubKeyB)
	exec := newValsetChangeExecutor(func() []pVals.Validator {
		return []pVals.Validator{{Address: addrB, VotingPower: 0}}
	})
	urequire.NoError(t, exec.Execute(cross), "second callback succeeds")

	// Effective view (still dirty=true) must now show only A.
	set := GetValidators()
	urequire.Equal(t, 1, len(set), "only A remains")
	urequire.Equal(t, pubKeyA, set[0].PubKey)
}

func TestCallback_AllZeroPowerPublishesEmpty(t *testing.T) {
	// All-remove proposal: every entry power=0. The callback does
	// not panic at the v3 boundary; it publishes entries=[]string{}
	// + dirty=true, leaving the chain-side liveCount==0 floor to
	// reject at EndBlocker (covered by app_test.go).
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	addrA := mustAddr(t, pubKeyA)
	exec := newValsetChangeExecutor(func() []pVals.Validator {
		return []pVals.Validator{{Address: addrA, VotingPower: 0}}
	})
	urequire.NoError(t, exec.Execute(cross), "all-remove callback does not panic")

	// dirty=true now; effective view = proposed = empty.
	urequire.Equal(t, 0, len(GetValidators()), "set drained")
	urequire.False(t, IsValidator(addrA))
}
