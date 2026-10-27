package blog

import (
	"testing"

	"gno.land/p/gnoland/blog/v0"
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm

// clearState wipes the global state between test calls
func clearState(t *testing.T) {
	t.Helper()

	b = &blog.Blog{
		Title:  "Gno.land's blog",
		Prefix: "/r/gnoland/blog:",
	}
	adminAddr = address("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx")
	inPause = false
	moderatorList = bptree.NewBPTree32()
	commenterList = bptree.NewBPTree32()
}

func TestBlog_AdminControls(cur realm, t *testing.T) {
	t.Run("non-admin call", func(cur realm, t *testing.T) {
		clearState(t)

		nonAdmin := testutils.TestAddress("bob")

		testing.SetOriginCaller(nonAdmin)
		testing.SetRealm(testing.NewUserRealm(nonAdmin))

		uassert.AbortsWithMessage(t, cur, errNotAdmin.Error(), func() {
			AdminSetInPause(cross(cur), true)
		})
	})

	t.Run("pause toggled", func(cur realm, t *testing.T) {
		clearState(t)

		testing.SetOriginCaller(adminAddr)
		testing.SetRealm(testing.NewUserRealm(adminAddr))

		uassert.NotAborts(t, cur, func() {
			AdminSetInPause(cross(cur), true)
		})

		uassert.True(t, inPause)
	})

	t.Run("admin set", func(cur realm, t *testing.T) {
		clearState(t)

		// Set the new admin
		var (
			oldAdmin = adminAddr
			newAdmin = testutils.TestAddress("alice")
		)

		testing.SetRealm(testing.NewUserRealm(adminAddr))
		AdminSetAdminAddr(cross(cur), newAdmin)

		uassert.Equal(t, adminAddr, newAdmin)

		// Make sure the old admin can't do anything
		testing.SetOriginCaller(oldAdmin)
		testing.SetRealm(testing.NewUserRealm(oldAdmin))

		uassert.AbortsWithMessage(t, cur, errNotAdmin.Error(), func() {
			AdminSetInPause(cross(cur), false)
		})
	})
}

func TestBlog_AddRemoveModerator(cur realm, t *testing.T) {
	clearState(t)

	mod := testutils.TestAddress("mod")

	// Add the moderator
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminAddModerator(cross(cur), mod)

	urequire.True(t, moderatorList.Has(mod.String()))

	// Remove the moderator
	AdminRemoveModerator(cross(cur), mod)

	// Make sure the moderator is disabled
	isMod := moderatorList.Get(mod.String())
	uassert.NotNil(t, isMod)

	uassert.False(t, isMod.(bool))
}

func TestBlog_AddCommenter(cur realm, t *testing.T) {
	clearState(t)

	var (
		mod       = testutils.TestAddress("mod")
		commenter = testutils.TestAddress("comm")
		rand      = testutils.TestAddress("rand")
	)

	// Appoint the moderator
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminAddModerator(cross(cur), mod)

	// Add a commenter as a mod
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddCommenter(cross(cur), commenter)

	uassert.True(t, commenterList.Has(commenter.String()))

	// Make sure a non-mod can't add commenters
	testing.SetOriginCaller(rand)
	testing.SetRealm(testing.NewUserRealm(rand))

	uassert.AbortsWithMessage(t, cur, errNotModerator.Error(), func() {
		ModAddCommenter(cross(cur), testutils.TestAddress("evil"))
	})

	// Remove a commenter
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModDelCommenter(cross(cur), commenter)

	active := commenterList.Get(commenter.String())
	uassert.False(t, active.(bool))
}

func TestBlog_ManagePost(cur realm, t *testing.T) {
	clearState(t)

	var (
		mod  = testutils.TestAddress("mod")
		slug = "slug"
	)

	// Appoint the moderator
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminAddModerator(cross(cur), mod)

	// Add the post
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddPost(
		cross(cur),
		slug, "title", "body", "2022-05-20T13:17:22Z", "moul", "tag",
	)

	// Make sure the post is present
	uassert.NotNil(t, b.GetPost(slug))

	// Remove the post
	ModRemovePost(cross(cur), slug)
	uassert.TypedNil(t, b.GetPost(slug))
}

func TestBlog_ManageComment(cur realm, t *testing.T) {
	clearState(t)

	var (
		slug = "slug"

		mod       = testutils.TestAddress("mod")
		commenter = testutils.TestAddress("comm")
	)

	// Appoint the moderator
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminAddModerator(cross(cur), mod)

	// Add a commenter as a mod
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddCommenter(cross(cur), commenter)

	uassert.True(t, commenterList.Has(commenter.String()))

	// Add the post
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddPost(
		cross(cur),
		slug, "title", "body", "2022-05-20T13:17:22Z", "moul", "tag",
	)

	// Make sure the post is present
	uassert.NotNil(t, b.GetPost(slug))

	// Add the comment
	testing.SetOriginCaller(commenter)
	testing.SetRealm(testing.NewUserRealm(commenter))
	uassert.NotAborts(t, cur, func() {
		AddComment(cross(cur), slug, "comment")
	})

	// Delete the comment
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	uassert.NotAborts(t, cur, func() {
		ModDelComment(cross(cur), slug, 0)
	})
}

// A call arriving through a realm is authenticated as that realm, so the
// roles held by whoever signed the transaction do not carry into it.
func TestBlog_CallerIsTheRealmNotTheSigner(cur realm, t *testing.T) {
	const intermediate = "gno.land/r/test/intermediate"

	t.Run("admin", func(cur realm, t *testing.T) {
		clearState(t)

		testing.SetOriginCaller(adminAddr)
		testing.SetRealm(testing.NewCodeRealm(intermediate))

		uassert.AbortsWithMessage(t, cur, errNotAdmin.Error(), func() {
			AdminSetInPause(cross(cur), true)
		})

		uassert.False(t, inPause)
	})

	t.Run("moderator", func(cur realm, t *testing.T) {
		clearState(t)

		mod := testutils.TestAddress("mod")

		testing.SetRealm(testing.NewUserRealm(adminAddr))
		AdminAddModerator(cross(cur), mod)

		testing.SetOriginCaller(mod)
		testing.SetRealm(testing.NewCodeRealm(intermediate))

		uassert.AbortsWithMessage(t, cur, errNotModerator.Error(), func() {
			ModAddPost(cross(cur), "slug", "title", "body", "2022-05-20T13:17:22Z", "moul", "tag")
		})

		uassert.TypedNil(t, b.GetPost("slug"))
	})

	t.Run("commenter", func(cur realm, t *testing.T) {
		clearState(t)

		commenter := testutils.TestAddress("comm")

		testing.SetRealm(testing.NewUserRealm(adminAddr))
		ModAddCommenter(cross(cur), commenter)
		ModAddPost(cross(cur), "slug", "title", "body", "2022-05-20T13:17:22Z", "moul", "tag")

		testing.SetOriginCaller(commenter)
		testing.SetRealm(testing.NewCodeRealm(intermediate))

		uassert.AbortsWithMessage(t, cur, errNotCommenter.Error(), func() {
			AddComment(cross(cur), "slug", "comment")
		})

		uassert.Equal(t, 0, b.GetPost("slug").Comments.Size())
	})
}
