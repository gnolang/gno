package impl

import (
	"testing"

	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao/v3/memberstore"
)

// Member addresses arrive as raw MsgCall strings (the VM does not validate an
// `address` arg), so a non-bech32 value — blank, oversized, or carrying
// markdown/pipe/HTML metachars — would otherwise be stored and then injected
// into the members render table. Both admission points reject it, matching the
// InitWithUsers precedent.
func TestNewAddMemberRequestRejectsInvalidAddress(cur realm, t *testing.T) {
	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))

	for _, bad := range []string{"evil|**pwn**", "", "not-an-address"} {
		urequire.PanicsWithMessage(t, cur,
			"invalid member address: "+bad,
			func() {
				NewAddMemberRequest(cur, address(bad), memberstore.T2, "portfolio")
			})
	}
}
