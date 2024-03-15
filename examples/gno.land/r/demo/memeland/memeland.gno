package memeland

import (
	memeland "gno.land/p/demo/staging/memeland/v1"
	"time"
)

var m = memeland.NewMemeland()

func PostMeme(data string, timestamp int64) string {
	return m.PostMeme(data, timestamp)
}

func Upvote(id string) string {
	return m.Upvote(id)
}

func GetPostsInRange(startTimestamp, endTimestamp int64, page, pageSize int, sortBy string) string {
	return m.GetPostsInRange(startTimestamp, endTimestamp, page, pageSize, sortBy)
}

func Render(path string) string {
	numOfMemes := int(m.MemeCounter)
	if numOfMemes == 0 {
		return "No memes posted yet! :/"
	}

	// Default render is get Posts since year 2000 to now
	return m.GetPostsInRange(0, time.Now().Unix(), 1, numOfMemes, "DATE_CREATED")
}
