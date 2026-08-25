package minisocial

import "gno.land/p/nt/ownable/v0"

// Unexported so no other realm can reach the ownable's mutators: an exported
// pointer would let any importing realm call TransferOwnership, and the write
// commits under this realm's authority.
var ownableInst = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn

// ResetPosts allows admin deletion of the posts
func ResetPosts(cur realm) {
	ownableInst.AssertOwnedBy(cur.Previous().Address())
	posts = nil
}
