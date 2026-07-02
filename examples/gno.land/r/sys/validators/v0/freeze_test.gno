// freeze_test.gno — basic tests for Phase A freeze/unfreeze
package validators

import (
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/urequire/v0"
)

func resetFreezeState(t *testing.T) {
	t.Helper()
	frozenSet = avl.NewTree()
	freezeAdmin = address("")
}

func TestIsFrozen_Default(t *testing.T) {
	resetFreezeState(t)
	resetValset(t)

	// No validators frozen initially.
	urequire.Equal(t, 0, len(GetFrozenValidators()))
}

func TestGetFreezeAdmin_Default(t *testing.T) {
	resetFreezeState(t)
	urequire.Equal(t, address("").String(), GetFreezeAdmin().String())
}

func TestFreezeAdmin_Roundtrip(t *testing.T) {
	resetFreezeState(t)

	// freezeAdmin starts empty.
	urequire.Equal(t, address("").String(), GetFreezeAdmin().String())

	// Directly set (bypassing GovDAO check) for unit testing purposes.
	freezeAdmin = address("g1someadminaddr")
	urequire.Equal(t, address("g1someadminaddr").String(), GetFreezeAdmin().String())

	// Reset.
	freezeAdmin = address("")
	urequire.Equal(t, address("").String(), GetFreezeAdmin().String())
}

func TestIsFrozen_AfterManualInsert(t *testing.T) {
	resetFreezeState(t)

	addr := address("g1testvalidator")
	urequire.False(t, IsFrozen(addr))

	// Manually insert to simulate a freeze (bypasses auth for unit testing).
	frozenSet.Set(addr.String(), frozenEntry{
		Addr:        addr,
		PubKey:      "pubkey",
		VotingPower: 10,
		Reason:      "test",
		FrozenAt:    1,
	})
	urequire.True(t, IsFrozen(addr))

	frozen := GetFrozenValidators()
	urequire.Equal(t, 1, len(frozen))
	urequire.Equal(t, addr.String(), frozen[0].String())
}
