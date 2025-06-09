package memeland

import (
	"std"
	"time"

	"gno.land/p/demo/memeland"
)

var m *memeland.Memeland

func init() {
	m = memeland.NewMemeland()
	m.TransferOwnership("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
}

func PostMeme(data string, timestamp int64) string {
	return m.PostMeme(data, timestamp)
}

func Upvote(id string) string {
	return m.Upvote(id)
}

func GetPostsInRange(startTimestamp, endTimestamp int64, page, pageSize int, sortBy string) string {
	return m.GetPostsInRange(startTimestamp, endTimestamp, page, pageSize, sortBy)
}

func RemovePost(id string) string {
	return m.RemovePost(id)
}

func GetOwner() std.Address {
	return m.Owner()
}

func TransferOwnership(newOwner std.Address) {
	if err := m.TransferOwnership(newOwner); err != nil {
		panic(err)
	}
}

func Render(path string) string {
	numOfMemes := int(m.MemeCounter)
	if numOfMemes == 0 {
		return "No memes posted yet! :/"
	}

	// Default render is get Posts since year 2000 to now
	return m.GetPostsInRange(0, time.Now().Unix(), 1, 10, "DATE_CREATED")
}
