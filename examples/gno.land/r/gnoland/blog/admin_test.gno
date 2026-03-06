package gnoblog

import (
	"testing"

	"gno.land/p/demo/blog"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// clearState wipes the global state between test calls
func clearState(t *testing.T) {
	t.Helper()

	b = &blog.Blog{
		Title:  "Gno.land's blog",
		Prefix: "/r/gnoland/blog:",
	}
	adminAddr = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	inPause = false
	moderatorList = avl.NewTree()
	commenterList = avl.NewTree()
}

func TestBlog_AdminControls(t *testing.T) {
	t.Run("non-admin call", func(t *testing.T) {
		clearState(t)

		nonAdmin := testutils.TestAddress("bob")

		testing.SetOriginCaller(nonAdmin)
		testing.SetRealm(testing.NewUserRealm(nonAdmin))

		uassert.AbortsWithMessage(t, errNotAdmin.Error(), func() {
			AdminSetInPause(cross, true)
		})
	})

	t.Run("pause toggled", func(t *testing.T) {
		clearState(t)

		testing.SetOriginCaller(adminAddr)
		testing.SetRealm(testing.NewUserRealm(adminAddr))

		uassert.NotAborts(t, func() {
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

		uassert.AbortsWithMessage(t, errNotAdmin.Error(), func() {
			AdminSetInPause(cross, false)
		})
	})
}

func TestBlog_AddRemoveModerator(t *testing.T) {
	clearState(t)

	mod := testutils.TestAddress("mod")

	// Add the moderator
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminAddModerator(cross, mod)

	_, found := moderatorList.Get(mod.String())
	urequire.True(t, found)

	// Remove the moderator
	AdminRemoveModerator(cross, mod)

	// Make sure the moderator is disabled
	isMod, _ := moderatorList.Get(mod.String())
	uassert.NotNil(t, isMod)

	uassert.False(t, isMod.(bool))
}

func TestBlog_AddCommenter(t *testing.T) {
	clearState(t)

	var (
		mod       = testutils.TestAddress("mod")
		commenter = testutils.TestAddress("comm")
		rand      = testutils.TestAddress("rand")
	)

	// Appoint the moderator
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminAddModerator(cross, mod)

	// Add a commenter as a mod
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddCommenter(cross, commenter)

	_, ok := commenterList.Get(commenter.String())
	uassert.True(t, ok)

	// Make sure a non-mod can't add commenters
	testing.SetOriginCaller(rand)
	testing.SetRealm(testing.NewUserRealm(rand))

	uassert.AbortsWithMessage(t, errNotModerator.Error(), func() {
		ModAddCommenter(cross, testutils.TestAddress("evil"))
	})

	// Remove a commenter
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModDelCommenter(cross, commenter)

	active, _ := commenterList.Get(commenter.String())
	uassert.False(t, active.(bool))
}

func TestBlog_ManagePost(t *testing.T) {
	clearState(t)

	var (
		mod  = testutils.TestAddress("mod")
		slug = "slug"
	)

	// Appoint the moderator
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminAddModerator(cross, mod)

	// Add the post
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddPost(
		cross,
		slug, "title", "body", "2022-05-20T13:17:22Z", "moul", "tag",
	)

	// Make sure the post is present
	uassert.NotNil(t, b.GetPost(slug))

	// Remove the post
	ModRemovePost(cross, slug)
	uassert.TypedNil(t, b.GetPost(slug))
}

func TestBlog_ManageComment(t *testing.T) {
	clearState(t)

	var (
		slug = "slug"

		mod       = testutils.TestAddress("mod")
		commenter = testutils.TestAddress("comm")
	)

	// Appoint the moderator
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	AdminAddModerator(cross, mod)

	// Add a commenter as a mod
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddCommenter(cross, commenter)

	_, ok := commenterList.Get(commenter.String())
	uassert.True(t, ok)

	// Add the post
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	ModAddPost(
		cross,
		slug, "title", "body", "2022-05-20T13:17:22Z", "moul", "tag",
	)

	// Make sure the post is present
	uassert.NotNil(t, b.GetPost(slug))

	// Add the comment
	testing.SetOriginCaller(commenter)
	testing.SetRealm(testing.NewUserRealm(commenter))
	uassert.NotAborts(t, func() {
		AddComment(cross, slug, "comment")
	})

	// Delete the comment
	testing.SetOriginCaller(mod)
	testing.SetRealm(testing.NewUserRealm(mod))
	uassert.NotAborts(t, func() {
		ModDelComment(cross, slug, 0)
	})
}
