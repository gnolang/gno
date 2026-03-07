package minisocial

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/seqid/v0"
)

var Ownable = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn

// ResetPosts allows admin deletion of the posts
func ResetPosts(_ realm) {
	Ownable.AssertOwnedByPrevious()
	posts = avl.NewTree()
	postID = seqid.ID(0)
	pag = pager.NewPager(posts, 5, true)
}
