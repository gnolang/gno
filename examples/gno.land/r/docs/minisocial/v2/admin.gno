package minisocial

import (
	"gno.land/p/nt/avl/pager/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/seqid/v0"
)

// owner gates admin-only operations. Ownable guards its own mutators, so
// exporting it would be safe; unexported keeps the admin surface to ResetPosts.
var owner = ownable.NewWithAddress("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx") // govdao t1 multisig

// ResetPosts allows admin deletion of the posts
func ResetPosts(cur realm) {
	owner.AssertOwnedBy(cur.Previous().Address())
	posts = avl.NewTree()
	postID = seqid.ID(0)
	pag = pager.NewPager(posts, 5, true)
}
