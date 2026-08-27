package minisocial

import "gno.land/p/nt/ownable/v0"

// Ownable guards its own mutators, so exporting this would be safe. It stays
// unexported to keep the realm's admin surface to ResetPosts alone.
var ownableInst = ownable.NewWithAddress("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh") // govdao t1 multisig

// ResetPosts allows admin deletion of the posts
func ResetPosts(cur realm) {
	ownableInst.AssertOwnedBy(cur.Previous().Address())
	posts = nil
}
