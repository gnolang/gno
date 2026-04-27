package storage

import (
	"gno.land/p/nt/avl/v0"
	"strconv"
)

func init() {
	// we write to three common data structure for persistence
	// avl.Tree, map and slice.
	posts0 := avl.NewTree()
	b0 := &TreeBoard{0, posts0}
	boards.Set(strconv.Itoa(0), b0)

	posts1 := make(map[int]Post)
	b1 := &MapBoard{1, posts1}
	boards.Set(strconv.Itoa(1), b1)

	posts2 := []Post{}
	b2 := &SliceBoard{2, posts2}
	boards.Set(strconv.Itoa(2), b2)
}

// post to all boards.
func AddPost(title, content string) {
	for i := 0; i < boards.Size(); i++ {
		boardId := strconv.Itoa(i)
		b, ok := boards.Get(boardId)
		if ok {
			b.(Board).AddPost(title, content)
		}
	}
}

func GetPost(boardId, postId int) string {
	b, ok := boards.Get(strconv.Itoa(boardId))
	var res string

	if ok {
		p, ok := b.(Board).GetPost(postId)
		if ok {
			res = p.title + "," + p.content
		}
	}
	return res
}

func GetPostSize(boardId int) int {
	b, ok := boards.Get(strconv.Itoa(boardId))
	var res int

	if ok {
		res = b.(Board).Size()
	} else {
		res = -1
	}

	return res
}

func GetBoardSize() int {
	return boards.Size()
}
