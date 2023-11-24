package postit

import (
	"strings"
	"time"
)

type Post struct {
	Id       uint64
	Username string
	Body     string

	Comments []Comment

	CreatedAt time.Time
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
