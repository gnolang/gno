package regspoofuser

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/r/demo/defi/foo20"
	"gno.land/r/nt/grc20reg/v0"
	"gno.land/r/tests/vm/regalias"
	"gno.land/r/tests/vm/regnoop"
	"gno.land/r/tests/vm/regspoof"
	"gno.land/r/tests/vm/regspoofapprove"
	"gno.land/r/tests/vm/regspooffrom"
	"gno.land/r/tests/vm/regzero"
)

const (
	thisRealm = "gno.land/r/tests/vm/regspoofuser"
	aliasKey  = "gno.land/r/tests/vm/regalias.ALI"
)

func TestForeignTransferRefused(cur realm, t *testing.T) {
	urequire.AbortsContains(t, cur, "must be declared in the registering realm", func() {
		regspoof.TryRegister(cross(cur))
	})
}

func TestForeignApproveRefused(cur realm, t *testing.T) {
	urequire.AbortsContains(t, cur, "must be declared in the registering realm", func() {
		regspoofapprove.TryRegister(cross(cur))
	})
}

func TestForeignTransferFromRefused(cur realm, t *testing.T) {
	urequire.AbortsContains(t, cur, "must be declared in the registering realm", func() {
		regspooffrom.TryRegister(cross(cur))
	})
}

func TestForeignApproveInTransferSlotRefused(cur realm, t *testing.T) {
	urequire.AbortsContains(t, cur, "must be declared in the registering realm", func() {
		regzero.TryRegister(cross(cur))
	})
}

// A token realm is trusted for its own token: its own functions may do nothing.
func TestOwnNoopFunctionsAccepted(cur realm, t *testing.T) {
	regnoop.TryRegister(cross(cur))
}

// After the registrant overwrites its token struct with foo20's, the key must
// still name the token that was registered.
func TestAliasedTokenKeepsItsLedger(cur realm, t *testing.T) {
	me := chain.PackageAddress(thisRealm)
	thief := testutils.TestAddress("thief")
	regalias.Swap(cross(cur))
	urequire.Equal(t, "ALI", grc20reg.MustGet(aliasKey).GetSymbol())
	grc20reg.Approve(0, cur, aliasKey, thief, 999)
	urequire.Equal(t, int64(0), foo20.Token.Allowance(me, thief))
}
