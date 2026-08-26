package minisocial

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/seqid/v0"
)

// owner gates admin-only operations. Realm globals are kept unexported so the
// ownable pointer (and its mutators) is reachable only through this realm's own
// gated entry points.
var owner = ownable.NewWithAddress("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh") // govdao t1 multisig

// ResetPosts allows admin deletion of the posts
func ResetPosts(cur realm) {
	owner.AssertOwnedBy(cur.Previous().Address())
	posts = avl.NewTree()
	postID = seqid.ID(0)
	pag = pager.NewPager(posts, 5, true)
}
