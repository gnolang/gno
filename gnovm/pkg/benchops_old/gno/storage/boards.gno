package storage

import (
	"gno.land/p/nt/avl"
	"strconv"
)

var boards avl.Tree

type Board interface {
	AddPost(title, content string)
	GetPost(id int) (Post, bool)
	Size() int
}

// posts are persisted in an avl tree
type TreeBoard struct {
	id    int
	posts *avl.Tree
}

func (b *TreeBoard) AddPost(title, content string) {
	n := b.posts.Size()
	p := Post{n, title, content}
	b.posts.Set(strconv.Itoa(n), p)
}

func (b *TreeBoard) GetPost(id int) (Post, bool) {
	p, ok := b.posts.Get(strconv.Itoa(id))
	if ok {
		return p.(Post), ok
	} else {
		return Post{}, ok
	}
}
func (b *TreeBoard) Size() int {
	return b.posts.Size()
}

// posts are persisted in a map
type MapBoard struct {
	id    int
	posts map[int]Post
}

func (b *MapBoard) AddPost(title, content string) {
	n := len(b.posts)
	p := Post{n, title, content}
	b.posts[n] = p
}

func (b *MapBoard) GetPost(id int) (Post, bool) {
	p, ok := b.posts[id]
	if ok {
		return p, ok
	} else {
		return Post{}, ok
	}
}
func (b *MapBoard) Size() int {
	return len(b.posts)
}

// posts are persisted in a slice
type SliceBoard struct {
	id    int
	posts []Post
}

func (b *SliceBoard) AddPost(title, content string) {
	n := len(b.posts)
	p := Post{n, title, content}
	b.posts = append(b.posts, p)
}

func (b *SliceBoard) GetPost(id int) (Post, bool) {
	if id < len(b.posts) {
		p := b.posts[id]

		return p, true
	} else {
		return Post{}, false
	}
}
func (b *SliceBoard) Size() int {
	return len(b.posts)
}

type Post struct {
	id      int
	title   string
	content string
}
