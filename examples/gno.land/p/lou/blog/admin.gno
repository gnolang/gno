package blog

import (
	"gno.land/p/demo/ownable"
)

func ResetPosts(admin ownable.Ownable) {
	admin.AssertOwnedByPrevious()
	// posts = avl.NewTree()
	// postID = seqid.ID(0)
	// pag = pager.NewPager(posts, 5, true)
}
