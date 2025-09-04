// XXX This is only used for testing in ./tests.
// Otherwise this package is deprecated.
// TODO: replace with a package that is supported, and delete this.

package dom

import (
	"strconv"

	"gno.land/p/nt/avl"
)

type Plot struct {
	Name     string
	Posts    avl.Tree // postsCtr -> *Post
	PostsCtr int
}

func (plot *Plot) AddPost(title string, body string) {
	ctr := plot.PostsCtr
	plot.PostsCtr++
	key := strconv.Itoa(ctr)
	post := &Post{
		Title: title,
		Body:  body,
	}
	plot.Posts.Set(key, post)
}

func (plot *Plot) String() string {
	str := "# [plot] " + plot.Name + "\n"
	if plot.Posts.Size() > 0 {
		plot.Posts.Iterate("", "", func(key string, value any) bool {
			str += "\n"
			str += value.(*Post).String()
			return false
		})
	}
	return str
}

type Post struct {
	Title    string
	Body     string
	Comments avl.Tree
}

func (post *Post) String() string {
	str := "## " + post.Title + "\n"
	str += ""
	str += post.Body
	if post.Comments.Size() > 0 {
		post.Comments.Iterate("", "", func(key string, value any) bool {
			str += "\n"
			str += value.(*Comment).String()
			return false
		})
	}
	return str
}

type Comment struct {
	Creator string
	Body    string
}

func (cmm Comment) String() string {
	return cmm.Body + " - @" + cmm.Creator + "\n"
}
