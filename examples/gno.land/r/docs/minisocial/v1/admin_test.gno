package minisocial

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// adminAddr mirrors the hardcoded owner of ownableInst in admin.gno.
const adminAddr = address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

func TestResetPostsUnauthorized(cur realm, t *testing.T) {
	posts = nil

	// A non-owner caller.
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("eve")))

	// Seed a post so we can confirm it survives the rejected reset.
	uassert.NoError(t, CreatePost(cross(cur), "Hello World!"))

	// The owner check must reject a non-owner: ResetPosts aborts.
	uassert.AbortsContains(t, cur, "ownable: caller is not owner", func(cur realm) {
		ResetPosts(cur)
	})

	uassert.True(t, len(posts) == 1, "posts must survive a rejected reset")
}

func TestResetPostsAuthorized(cur realm, t *testing.T) {
	posts = nil

	// The owner caller.
	testing.SetRealm(testing.NewUserRealm(adminAddr))

	uassert.NoError(t, CreatePost(cross(cur), "Hello World!"))
	uassert.True(t, len(posts) == 1, "expected one post before reset")

	ResetPosts(cross(cur))

	uassert.True(t, len(posts) == 0, "expected posts cleared after authorized reset")
}
