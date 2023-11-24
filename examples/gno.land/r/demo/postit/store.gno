package postit

import (
	"gno.land/p/demo/avl"
)

type Trending struct {
	Tag   string
	Count uint64
}

var (
	postCount uint64
	userCount uint64

	trending [10]Trending

	hashtags          avl.Tree // hashtag -> count
	users             avl.Tree // username -> User
	addressToUsername avl.Tree // std.Address -> username
	posts             avl.Tree // id -> Post
)

func setPostCount(count uint64) {
	postCount = count
}

func updateTrending(tag string, count uint64) {
	for i := 0; i < len(trending); i++ {
		if trending[i].Count <= count {
			trending[i].Tag = tag
			trending[i].Count = count
		}
	}
}
