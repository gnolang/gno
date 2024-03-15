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

func GetPostsInRange(startTimestamp, endTimestamp int64, page, pageSize int, sortBy string) string {
	if m == nil {
		return "memeland does not exist"
	}

	return m.GetPostsInRange(startTimestamp, endTimestamp, page, pageSize, sortBy)
}

func Render(path string) string {
	numOfMemes := int(m.MemeCounter)
	if numOfMemes == 0 {
		return "[]"
	}

	// Default render is get Posts since year 2000 to now
	return m.GetPostsInRange(952733931, time.Now().Unix(), 1, numOfMemes, "DATE_CREATED")
}
