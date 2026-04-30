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

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
	pVals "gno.land/p/sys/validators"
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

func TestCallback_AccumulatesOnDirtyProposed_SecondRemoves(t *testing.T) {
	// Same-block accumulation, remove path: a prior callback wrote
	// valset:proposed + dirty=true with B added. A second callback
	// removes B. Effective view post-second must show only A.
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})
	testing.SetSysParamStrings(module, submodule, propKey, []string{pubKeyA + ":10", pubKeyB + ":1"})
	testing.SetSysParamBool(module, submodule, dirtyKey, true)

	addrB := mustAddr(t, pubKeyB)
	exec := newValsetChangeExecutor(func() []pVals.Validator {
		return []pVals.Validator{{Address: addrB, VotingPower: 0}}
	})
	urequire.NoError(t, exec.Execute(cross), "second callback succeeds")

	set := GetValidators()
	urequire.Equal(t, 1, len(set), "only A remains")
	urequire.Equal(t, pubKeyA, set[0].PubKey)
}

func TestCallback_AccumulatesOnDirtyProposed_SecondAdds(t *testing.T) {
	// Same-block accumulation, add path: this is the regression test
	// the remove-only case can't catch — if the second callback
	// clobbered proposed instead of accumulating, it would publish
	// [A,C] (losing B) instead of [A,B,C].
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})
	testing.SetSysParamStrings(module, submodule, propKey, []string{pubKeyA + ":10", pubKeyB + ":1"})
	testing.SetSysParamBool(module, submodule, dirtyKey, true)

	addrC := mustAddr(t, pubKeyC)
	exec := newValsetChangeExecutor(func() []pVals.Validator {
		return []pVals.Validator{{Address: addrC, PubKey: pubKeyC, VotingPower: 2}}
	})
	urequire.NoError(t, exec.Execute(cross), "second callback succeeds")

	// Effective view must contain all three: A from current, B from
	// the prior callback's proposed, and C from this callback.
	set := GetValidators()
	urequire.Equal(t, 3, len(set), "A, B, and C all present")
	seen := map[string]bool{}
	for _, v := range set {
		seen[v.PubKey] = true
	}
	urequire.True(t, seen[pubKeyA], "A retained")
	urequire.True(t, seen[pubKeyB], "B retained from prior callback")
	urequire.True(t, seen[pubKeyC], "C added by second callback")
}

func TestCallback_PanicsOnRemoveOfMissing(t *testing.T) {
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	addrC := mustAddr(t, pubKeyC) // not in set
	exec := newValsetChangeExecutor(func() []pVals.Validator {
		return []pVals.Validator{{Address: addrC, VotingPower: 0}}
	})
	urequire.AbortsWithMessage(t,
		"validator does not exist: "+addrC.String(),
		func() { exec.Execute(cross) })
}

func TestCallback_PanicsOnAddOfExisting(t *testing.T) {
	resetValset(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	addrA := mustAddr(t, pubKeyA) // already in set
	exec := newValsetChangeExecutor(func() []pVals.Validator {
		return []pVals.Validator{{Address: addrA, PubKey: pubKeyA, VotingPower: 5}}
	})
	urequire.AbortsWithMessage(t,
		"validator already exists: "+addrA.String(),
		func() { exec.Execute(cross) })
}

func TestNewProposalRequest_RejectsAddressPubKeyMismatch(t *testing.T) {
	// Adversarial proposal: c.Address is one identity but c.PubKey
	// derives to another. Without the validation the proposal's
	// description would advertise the proposer-supplied address,
	// while the callback would publish PubKey only and consensus
	// would apply the pubkey-derived owner — voters mislead. The
	// validation rejects at creation, before voting.
	resetValset(t)

	addrA := mustAddr(t, pubKeyA)
	derivedB := mustAddr(t, pubKeyB)

	urequire.PanicsWithMessage(t,
		"address/pubkey mismatch: address="+addrA.String()+" pubkey-derived="+derivedB.String(),
		func() {
			_ = NewProposalRequest(func() []pVals.Validator {
				return []pVals.Validator{{Address: addrA, PubKey: pubKeyB, VotingPower: 1}}
			}, "title", "")
		})
}

func TestNewProposalRequest_RejectsInvalidPubKey(t *testing.T) {
	resetValset(t)

	uassert.PanicsContains(t, "invalid pubkey:", func() {
		_ = NewProposalRequest(func() []pVals.Validator {
			return []pVals.Validator{{Address: mustAddr(t, pubKeyA), PubKey: "not-a-pubkey", VotingPower: 1}}
		}, "title", "")
	})
}

func TestNewProposalRequest_AcceptsConsistentAddressAndPubKey(t *testing.T) {
	// Sanity: a proposer that supplies a pubkey-derived address
	// passes validation. Mirrors what valopers/proposal/proposal.gno
	// does in production.
	resetValset(t)

	addrA := mustAddr(t, pubKeyA)
	urequire.NotPanics(t, func() {
		_ = NewProposalRequest(func() []pVals.Validator {
			return []pVals.Validator{{Address: addrA, PubKey: pubKeyA, VotingPower: 1}}
		}, "title", "")
	})
}

func TestNewProposalRequest_RemoveSkipsPubKeyCheck(t *testing.T) {
	// Removes carry no PubKey. The validation must not require one
	// for power=0 entries; the lookup-and-panic happens at execution
	// time against the effective set.
	resetValset(t)

	addrA := mustAddr(t, pubKeyA)
	urequire.NotPanics(t, func() {
		// Empty PubKey on a power=0 entry: must build the request
		// without firing "invalid pubkey".
		_ = NewProposalRequest(func() []pVals.Validator {
			return []pVals.Validator{{Address: addrA, VotingPower: 0}}
		}, "title", "")
	})
}

// Empty-final-set rejection in the callback (per D4ryl00 review,
// validators.gno:147) is intentionally not tested at the gno layer:
// the panic happens through dao.Executor.Execute which is a cross-
// realm call, and gno cross-realm panics abort instead of being
// recoverable — uassert.PanicsContains only catches same-realm
// panics. The Go-side TestEndBlocker_RejectsEmptyValset (app_test.go)
// covers the chain's own empty-set rejection on the read path.
