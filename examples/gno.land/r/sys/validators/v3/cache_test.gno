package validators

import (
	"testing"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
	sysparams "gno.land/r/sys/params"
)

// resetCache clears valoperCache so subtests don't leak state.
func resetCache() {
	valoperCache = bptree.NewBPTree32()
}

func TestNotifyValoperChanged_HappyPath(t *testing.T) {
	resetCache()

	op := testutils.TestAddress("op-A")
	signingAddr := mustAddr(t, pubKeyA)

	// Caller realm = valopers; auth check passes.
	testing.SetRealm(testing.NewCodeRealm(valopersRealmPath))

	NotifyValoperChanged(cross, op, pubKeyA, signingAddr, true)

	rawEntry, ok := valoperCache.Get(op.String())
	urequire.True(t, ok, "cache entry must exist")
	entry := rawEntry.(cacheEntry)
	uassert.Equal(t, pubKeyA, entry.SigningPubKey)
	uassert.Equal(t, signingAddr, entry.SigningAddress)
	uassert.Equal(t, true, entry.KeepRunning)

	// Subsequent call updates the same slot.
	signingAddrB := mustAddr(t, pubKeyB)
	NotifyValoperChanged(cross, op, pubKeyB, signingAddrB, false)

	rawEntry, ok = valoperCache.Get(op.String())
	urequire.True(t, ok)
	entry = rawEntry.(cacheEntry)
	uassert.Equal(t, pubKeyB, entry.SigningPubKey)
	uassert.Equal(t, signingAddrB, entry.SigningAddress)
	uassert.Equal(t, false, entry.KeepRunning)
}

func TestNotifyValoperChanged_RejectsNonValopersCaller(t *testing.T) {
	resetCache()

	op := testutils.TestAddress("op-A")
	signingAddr := mustAddr(t, pubKeyA)

	// Caller realm = some other realm; auth check rejects.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/attacker"))

	uassert.AbortsContains(t, "caller realm must be "+valopersRealmPath, func() {
		NotifyValoperChanged(cross, op, pubKeyA, signingAddr, true)
	})

	// User MsgCall — UserRealm has empty pkgpath; rejected the same way.
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("user")))

	uassert.AbortsContains(t, "caller realm must be "+valopersRealmPath, func() {
		NotifyValoperChanged(cross, op, pubKeyA, signingAddr, true)
	})
}

func TestRotateValoperSigningKey_AppliesRotation(t *testing.T) {
	resetValset(t)
	resetCache()

	// Seed the valset with op-A signing under pubKeyA at power=10.
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	op := testutils.TestAddress("op-A")
	testing.SetRealm(testing.NewCodeRealm(valopersRealmPath))

	RotateValoperSigningKey(cross, op, pubKeyA, pubKeyB)

	// After rotation the effective set (proposed-when-dirty) should
	// contain pubKeyB at the same power; pubKeyA should be gone.
	uassert.True(t, sysparams.ValsetDirty(), "dirty bit set after RotateValoperSigningKey")

	effective := sysparams.GetValsetEffective()
	urequire.Equal(t, 1, len(effective))
	uassert.Equal(t, pubKeyB, effective[0].PubKey)
	uassert.Equal(t, uint64(10), effective[0].VotingPower)
}

func TestRotateValoperSigningKey_OperatorNotInValset_NoOp(t *testing.T) {
	resetValset(t)
	resetCache()

	// Empty valset; the operator is not currently signing.
	op := testutils.TestAddress("op-A")
	testing.SetRealm(testing.NewCodeRealm(valopersRealmPath))

	// Should not panic; no sysparams write expected (dirty stays false).
	uassert.NotAborts(t, func() {
		RotateValoperSigningKey(cross, op, pubKeyA, pubKeyB)
	})

	uassert.False(t, sysparams.ValsetDirty(), "dirty bit must remain false on no-op rotation")
	uassert.Equal(t, 0, len(sysparams.GetValsetEffective()))
}

func TestRotateValoperSigningKey_RejectsNonValopersCaller(t *testing.T) {
	resetValset(t)
	resetCache()
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	op := testutils.TestAddress("op-A")

	// Foreign realm — rejected.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/attacker"))
	uassert.AbortsContains(t, "caller realm must be "+valopersRealmPath, func() {
		RotateValoperSigningKey(cross, op, pubKeyA, pubKeyB)
	})

	// User MsgCall — UserRealm pkgpath is "" — rejected.
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("user")))
	uassert.AbortsContains(t, "caller realm must be "+valopersRealmPath, func() {
		RotateValoperSigningKey(cross, op, pubKeyA, pubKeyB)
	})

	// Sanity: the valset wasn't mutated by the rejected calls (dirty
	// bit stays false; effective set still equals the seeded current).
	uassert.False(t, sysparams.ValsetDirty())
	effective := sysparams.GetValsetEffective()
	urequire.Equal(t, 1, len(effective))
	uassert.Equal(t, pubKeyA, effective[0].PubKey)
}

func TestAssertGenesisValopersConsistent_HappyPath(t *testing.T) {
	resetValset(t)
	resetCache()
	testing.SetHeight(0) // assertion is genesis-mode only

	// Seed valset:current with two genesis validators.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10",
		pubKeyB + ":5",
	})

	// Seed valoperCache with profiles whose SigningAddress matches.
	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
	})

	uassert.NotAborts(t, func() {
		AssertGenesisValopersConsistent(cross)
	})
}

func TestAssertGenesisValopersConsistent_PanicsOnMissingProfile(t *testing.T) {
	resetValset(t)
	resetCache()
	testing.SetHeight(0)

	// Two validators in valset:current but only one has a profile.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10",
		pubKeyB + ":5",
	})

	opA := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		// op for pubKeyB intentionally missing.
	})

	uassert.AbortsContains(t, "no corresponding valoper profile", func() {
		AssertGenesisValopersConsistent(cross)
	})
}

func TestAssertGenesisValopersConsistent_EmptyValsetTrivial(t *testing.T) {
	resetValset(t)
	resetCache()
	testing.SetHeight(0)

	// Empty valset → assertion trivially holds (no entries to check).
	uassert.NotAborts(t, func() {
		AssertGenesisValopersConsistent(cross)
	})
}

func TestAssertGenesisValopersConsistent_RejectsPostGenesis(t *testing.T) {
	resetValset(t)
	resetCache()
	// Default test height is non-zero (123); the assertion must
	// refuse to run outside genesis-mode replay.
	uassert.AbortsContains(t, "only callable during genesis-mode replay", func() {
		AssertGenesisValopersConsistent(cross)
	})
}

func TestRotateValoperSigningKey_AccumulatesAcrossSameBlock(t *testing.T) {
	resetValset(t)
	resetCache()

	// Seed valset with two operators signing under pubKeyA and pubKeyC.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10",
		pubKeyC + ":5",
	})

	opA := testutils.TestAddress("op-A")
	opC := testutils.TestAddress("op-C")
	testing.SetRealm(testing.NewCodeRealm(valopersRealmPath))

	// First rotation: A → B. Read-modify-write of GetValsetEffective.
	RotateValoperSigningKey(cross, opA, pubKeyA, pubKeyB)

	// Second rotation in the same block: C → A. Should not clobber
	// the first rotation; should read proposed-when-dirty as baseline.
	RotateValoperSigningKey(cross, opC, pubKeyC, pubKeyA)

	// Final effective set: {pubKeyB:10, pubKeyA:5}. Order may differ
	// because GetValsetEffective parses sysparams' sorted-string slot
	// and returns []Validator preserving that order.
	effective := sysparams.GetValsetEffective()
	urequire.Equal(t, 2, len(effective))

	// Build a power-by-pubkey map and assert.
	powerOf := map[string]uint64{}
	for _, v := range effective {
		powerOf[v.PubKey] = v.VotingPower
	}
	uassert.Equal(t, uint64(10), powerOf[pubKeyB])
	uassert.Equal(t, uint64(5), powerOf[pubKeyA])
	_, ok := powerOf[pubKeyC]
	uassert.False(t, ok, "pubKeyC should be replaced by pubKeyA at the rotated slot")
}
