package minisocial

import (
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/seqid"
)

var Ownable = ownable.NewWithAddressByPrevious("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn

// ResetPosts allows admin deletion of the posts
func ResetPosts(_ realm) {
	Ownable.AssertOwned()
	posts = avl.NewTree()
	postID = seqid.ID(0)
	pag = pager.NewPager(posts, 5, true)
}
