package memeland

import (
	"gno.land/p/demo/avl"
	"std"
	"time"
)

type Feed struct {
	Memes     avl.Tree
	MemeCount uint64
	Users     avl.Tree // userID > addr
}

type Post struct {
	user       std.Address // maybe, if user is registered in r/demo/users, fetch username
	meme       string
	timePosted time.Time
	upvotes    uint64
	downvotes  uint64
	// todo add comments
}

func NewPost(meme, timePosted string) (*Post, error) {
	if meme == "" || timePosted == "" {
		return nil, ErrInvalidInput
	}

	parsedTime, err := time.Parse(timePosted, time.RFC3339)
	if err != nil {
		return nil, ErrInvalidTime
	}

	return &Post{
		user:       std.GetOrigCaller(),
		meme:       meme,
		timePosted: parsedTime,
	}, nil
}
