package minisocial

import "gno.land/p/demo/ownable"

var Ownable = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn

// ResetPosts allows admin deletion of the posts
func ResetPosts() {
	crossing()
	Ownable.AssertOwnedByPrevious()
	posts = nil
}
