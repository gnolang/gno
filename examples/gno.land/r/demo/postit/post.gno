package postit

import (
	"strings"
	"time"
)

type Post struct {
	Id         uint64
	Username   string
	Body       string
	Attachment string // base64

	Comments []Comment

	CreatedAt time.Time
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
