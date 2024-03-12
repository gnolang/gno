package memeland

import (
	"gno.land/p/demo/memeland"
	"time"
)

var m = memeland.NewMemeland()

func PostMeme(data string, timestamp int64) string {
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

func GetPostsInRange(startTimestamp, endTimestamp int64, page, pageSize int) string {
	if m == nil {
		return "memeland does not exist"
	}

	return m.GetPostsInRange(startTimestamp, endTimestamp, page, pageSize)
}

func Render(path string) string {
	// Default render is get Posts since year 2000 to now
	postsJson := m.GetPostsInRange(952733931, time.Now().Unix(), 1, 100)
	if postsJson == "[]" {
		return "no memes yet :c"
	}
	return postsJson
}
