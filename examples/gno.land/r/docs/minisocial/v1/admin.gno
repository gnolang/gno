package minisocial

import "gno.land/p/nt/ownable/v0"

// Unexported so no other realm can reach the ownable's mutators: an exported
// pointer would let any importing realm call TransferOwnership, and the write
// commits under this realm's authority.
var ownableInst = ownable.NewWithAddress("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh") // govdao t1 multisig

// ResetPosts allows admin deletion of the posts
func ResetPosts(cur realm) {
	ownableInst.AssertOwnedBy(cur.Previous().Address())
	posts = nil
}
