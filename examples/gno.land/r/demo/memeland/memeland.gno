package memeland

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/memeland"
	"strconv"
	"time"
)

var m = &memeland.Memeland{
	Posts: avl.NewTree(),
}

func PostMeme(data, timestamp string) string {
	if m == nil {
		return "memeland does not exist"
	}
	return m.PostMeme(data, timestamp)
}

func Upvote(id string) string {
	if m == nil {
		return "memeland does not exist"
	}
	return m.Upvote(id)
}

func GetPostsByTimestamp(startTimestamp, endTimestamp string, page, pageSize int) string {
	if m == nil {
		return "memeland does not exist"
	}

	return m.GetPostsByTimestamp(startTimestamp, endTimestamp, page, pageSize)
}

func Render(path string) string {
	postsJson := m.GetPostsByTimestamp("952733931", strconv.Itoa(int(time.Now().Unix())), 1, 100)
	if postsJson == "[]" {
		return "no memes yet :c"
	}
	return postsJson
}
