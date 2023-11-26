package postit

import (
	"gno.land/p/demo/avl"
)

var (
	postCount uint64
	userCount uint64

	feed     []uint64
	trending [10]Trending

	hashtags          avl.Tree // hashtag -> count
	users             avl.Tree // username -> User
	addressToUsername avl.Tree // std.Address -> username
	posts             avl.Tree // id -> Post
)
