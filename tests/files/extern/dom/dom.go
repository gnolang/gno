package dom

import (
	"strconv"

	"github.com/gnolang/gno/_test/avl"
)

type Plot struct {
	Name     string
	Posts    *avl.Tree // postsCtr -> *Post
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
	posts2, _ := plot.Posts.Set(key, post)
	plot.Posts = posts2
}

func (plot *Plot) String() string {
	str := "# [plot] " + plot.Name + "\n"
	if plot.Posts.Size() > 0 {
		plot.Posts.Traverse(true, func(n *avl.Tree) bool {
			str += "\n"
			str += n.Value().(*Post).String()
			return false
		})
	}
	return str
}

type Post struct {
	Title    string
	Body     string
	Comments *avl.Tree
}

func (post *Post) String() string {
	str := "## " + post.Title + "\n"
	str += ""
	str += post.Body
	if post.Comments.Size() > 0 {
		post.Comments.Traverse(true, func(n *avl.Tree) bool {
			str += "\n"
			str += n.Value().(*Comment).String()
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
