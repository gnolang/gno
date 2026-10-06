package microposts

import (
	"strconv"
	"time"
)

var posts []*Post

func CreatePost(cur realm, text string) {
	posts = append(posts, &Post{
		text:      text,
		author:    cur.Previous().Address(), // provided by env
		createdAt: time.Now(),
	})
}

func Render(_ string) string {
	out := "# Posts\n"
	for i := len(posts) - 1; i >= 0; i-- {
		out += "### Post " + strconv.Itoa(i) + "\n" + posts[i].String()
	}
	return out
}
