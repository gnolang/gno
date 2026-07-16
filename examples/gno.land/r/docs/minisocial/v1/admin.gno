package minisocial

import "gno.land/p/nt/ownable/v0"

// Unexported so no other realm can reassign the binding and hijack admin rights;
// an exported package var is writable by anyone with the import.
var ownableInst = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn

// ResetPosts allows admin deletion of the posts
func ResetPosts(cur realm) {
	ownableInst.AssertOwnedBy(cur.Previous().Address())
	posts = nil
}
