package pages

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// A call arriving through a realm is authenticated as that realm, so the
// roles held by whoever signed the transaction do not carry into it.
func TestPages_CallerIsTheRealmNotTheSigner(cur realm, t *testing.T) {
	const intermediate = "gno.land/r/test/intermediate"

	t.Run("admin", func(cur realm, t *testing.T) {
		testing.SetOriginCaller(adminAddr)
		testing.SetRealm(testing.NewCodeRealm(intermediate))

		uassert.AbortsWithMessage(t, cur, "access restricted.", func() {
			AdminSetInPause(cross(cur), true)
		})

		uassert.False(t, inPause)

		testing.SetRealm(testing.NewUserRealm(adminAddr))
		AdminSetInPause(cross(cur), true)
		uassert.True(t, inPause)
		AdminSetInPause(cross(cur), false)
	})

	t.Run("moderator", func(cur realm, t *testing.T) {
		mod := testutils.TestAddress("mod")

		testing.SetRealm(testing.NewUserRealm(adminAddr))
		AdminAddModerator(cross(cur), mod)

		testing.SetOriginCaller(mod)
		testing.SetRealm(testing.NewCodeRealm(intermediate))

		uassert.AbortsWithMessage(t, cur, "access restricted", func() {
			ModAddPost(cross(cur), "slug", "title", "body", "2026-01-01T00:00:00Z", "author", "tag")
		})

		uassert.TypedNil(t, b.GetPost("slug"))

		testing.SetRealm(testing.NewUserRealm(mod))
		ModAddPost(cross(cur), "slug", "title", "body", "2026-01-01T00:00:00Z", "author", "tag")
		uassert.NotTypedNil(t, b.GetPost("slug"))
		b.RemovePost("slug")
	})
}
