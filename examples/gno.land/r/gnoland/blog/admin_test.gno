package gnoblog

import (
	"testing"

	"gno.land/p/demo/blog"
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
	adminAddr = address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh")
	inPause = false
	moderatorList = bptree.NewBPTree32()
	commenterList = bptree.NewBPTree32()
}

func TestBlog_AdminControls(cur realm, t *testing.T) {
	t.Run("non-admin call", func(t *testing.T) {
		clearState(t)

		nonAdmin := testutils.TestAddress("bob")

		testing.SetOriginCaller(nonAdmin)
		testing.SetRealm(testing.NewUserRealm(nonAdmin))

		uassert.AbortsWithMessage(t, cur, errNotAdmin.Error(), func() {
			AdminSetInPause(cross, true)
		})
	})

	t.Run("pause toggled", func(t *testing.T) {
		clearState(t)

		testing.SetOriginCaller(adminAddr)
		testing.SetRealm(testing.NewUserRealm(adminAddr))

		uassert.NotAborts(t, cur, func() {
			AdminSetInPause(cross, true)
		})

		uassert.True(t, inPause)
	})

	t.Run("admin set", func(t *testing.T) {
		clearState(t)

		// Set the new admin
		var (
			oldAdmin = adminAddr
			newAdmin = testutils.TestAddress("alice")
		)

		testing.SetRealm(testing.NewUserRealm(adminAddr))
		AdminSetAdminAddr(cross, newAdmin)

		uassert.Equal(t, adminAddr, newAdmin)

		// Make sure the old admin can't do anything
		testing.SetOriginCaller(oldAdmin)
		testing.SetRealm(testing.NewUserRealm(oldAdmin))

		uassert.AbortsWithMessage(t, cur, errNotAdmin.Error(), func() {
			AdminSetInPause(cross, false)
		})
	})
}

func TestBlog_AddRemoveModerator(cur realm, t *testing.T) {
	clearState(t)

	mod := testutils.TestAddress("mod")

	// Add the moderator
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminAddModerator(cross2(cur), mod)

	_, found := moderatorList.Get(mod.String())
	urequire.True(t, found)

	// Remove the moderator
	AdminRemoveModerator(cross2(cur), mod)

	// Make sure the moderator is disabled
	isMod, _ := moderatorList.Get(mod.String())
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
	AdminAddModerator(cross2(cur), mod)

	// Add a commenter as a mod
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddCommenter(cross2(cur), commenter)

	_, ok := commenterList.Get(commenter.String())
	uassert.True(t, ok)

	// Make sure a non-mod can't add commenters
	testing.SetOriginCaller(rand)
	testing.SetRealm(testing.NewUserRealm(rand))

	uassert.AbortsWithMessage(t, cur, errNotModerator.Error(), func() {
		ModAddCommenter(cross2(cur), testutils.TestAddress("evil"))
	})

	// Remove a commenter
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModDelCommenter(cross2(cur), commenter)

	active, _ := commenterList.Get(commenter.String())
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
	AdminAddModerator(cross2(cur), mod)

	// Add the post
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddPost(
		cross2(cur),
		slug, "title", "body", "2022-05-20T13:17:22Z", "moul", "tag",
	)

	// Make sure the post is present
	uassert.NotNil(t, b.GetPost(slug))

	// Remove the post
	ModRemovePost(cross2(cur), slug)
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
	AdminAddModerator(cross2(cur), mod)

	// Add a commenter as a mod
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddCommenter(cross2(cur), commenter)

	_, ok := commenterList.Get(commenter.String())
	uassert.True(t, ok)

	// Add the post
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddPost(
		cross2(cur),
		slug, "title", "body", "2022-05-20T13:17:22Z", "moul", "tag",
	)

	// Make sure the post is present
	uassert.NotNil(t, b.GetPost(slug))

	// Add the comment
	testing.SetOriginCaller(commenter)
	testing.SetRealm(testing.NewUserRealm(commenter))
	uassert.NotAborts(t, cur, func() {
		AddComment(cross2(cur), slug, "comment")
	})

	// Delete the comment
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	uassert.NotAborts(t, cur, func() {
		ModDelComment(cross2(cur), slug, 0)
	})
}
