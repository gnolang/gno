package minisocial

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/seqid/v0"
)

// owner gates admin-only operations. Ownable guards its own mutators, so
// exporting it would be safe; unexported keeps the admin surface to ResetPosts.
var owner = ownable.NewWithAddress("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh") // govdao t1 multisig

// ResetPosts allows admin deletion of the posts
func ResetPosts(cur realm) {
	owner.AssertOwnedBy(cur.Previous().Address())
	posts = avl.NewTree()
	postID = seqid.ID(0)
	pag = pager.NewPager(posts, 5, true)
}
