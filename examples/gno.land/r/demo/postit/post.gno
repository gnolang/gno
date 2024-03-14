package postit

import (
	"strconv"
	"strings"
	"time"
)

type Post struct {
	Id         uint64
	Username   string
	Body       string
	Attachment string // base64

	LikedBy  []string
	Comments []Comment

	CreatedAt time.Time
}

func getPostRef(id uint64) *Post {
	res, ok := posts.Get(strconv.FormatUint(id, 10))
	if !ok {
		panic("post not found")
	}
	return res.(*Post)
}

func setPostCount(count uint64) {
	postCount = count
}

func parsehashTags(b string) (res []string) {
	words := strings.Fields(b)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			res = append(res, word[1:])
		}
	}
	return
}
