package groups

import (
	"strconv"

	//"github.com/gnolang/gno/examples/gno.land/p/avl"
	"gno.land/p/avl"
)

//----------------------------------------
// Group

type Group struct {
	Name     string
	Posts    *avl.Tree // postsCtr -> *Post
	PostsCtr uint64
}

func (group *Group) AddPost(title string, body string) *Post {
	ctr := group.PostsCtr
	group.PostsCtr++
	key := strconv.Itoa(int(ctr)) // TODO fix
	post := &Post{
		ID:    ctr,
		Title: title,
		Body:  body,
	}
	posts2, _ := group.Posts.Set(key, post)
	group.Posts = posts2
	return post
}

func (group *Group) String() string {
	str := "# [group] " + group.Name + "\n"
	if group.Posts.Size() > 0 {
		group.Posts.Traverse(true, func(n *avl.Tree) bool {
			str += "\n"
			str += n.Value().(*Post).String()
			return false
		})
	}
	return str
}

//----------------------------------------
// Post & Comment

type Post struct {
	ID       uint64
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
