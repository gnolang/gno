package ownable

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func TestChangeOwnerRejectsInvalidAddress(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("alice")))
	InitOwner(cross(cur))

	uassert.AbortsWithMessage(t, cur, "invalid owner address", func() {
		ChangeOwner(cross(cur), address("**[click](https://evil.example)**"))
	})
}
