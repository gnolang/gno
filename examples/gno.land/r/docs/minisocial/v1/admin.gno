package minisocial

import "gno.land/p/nt/ownable/v0"

var Ownable = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn

// ResetPosts allows admin deletion of the posts
func ResetPosts(_ realm) {
	Ownable.AssertOwnedByPrevious()
	posts = nil
}
