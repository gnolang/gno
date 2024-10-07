package boardsv2

import (
	"errors"
	"strconv"
)

// NOTE: Boards type should be a realm type
type Boards struct {
	// NOTE: Most of all types could be saved within the same AVL tree, prefixing the level
	// NOTE: We might want different AVL trees to avoid using level prefixes
	// NOTE: We could introduce the stores concept for the posts type instead
	// NOTE: Again, we could consider using Node instead of Post, and tree instead of posts (semantics)
	posts avl.Tree // string(Post.Level + slug) -> *Post (post, comment, poll)
}

// NOTE: Iterating only Post type is confusing semantically
// NOTE: Potential need cast/switch/type-check when iterating.
func (b Boards) Iterate(path string, fn func(*Post) bool) bool         {}
func (b Boards) IteratePosts(path string, fn func(*Post) bool) bool    {}
func (b Boards) IterateComments(path string, fn func(*Post) bool) bool {}

// How to map render paths to actual post instances?
//
// AVL KEYS BY LEVEL PREFIX (start/end)
// Boards => 0_ ... 1_
// Posts => 1_BOARD/ ... 2_
// Comments => 2_BOARD/POST/ ... 3_
//
// HOW TO GUESS PREFIX FROM SLUG
// User enters a SLUG => (one part => 1_BOARD)(more than one part => 1_BOARD/POST)
// How to recognize comments? Should be URL accesible? We could use ":" as separator (not optimal)

func (b *Boards) Set(p *Post) (updated bool) {
	key := newKey(p.Level, p.Slug())
	return b.posts.Set(key, p)
}

func (b *Boards) Remove(level int, path string) (_ *Post, removed bool) {
	key := newKey(level, path)
	if v, removed := b.posts.Remove(key); removed {
		return v.(*Post), true
	}
	return nil, false
}

func (b Boards) Get(level int, path string) (_ *Post, found bool) {
	key := newKey(level, path)
	if v, found := b.posts.Get(key); found {
		return v.(*Post), true
	}
	return "", false
}

func newKey(level int, path string) string {
	return strconv.Itoa(level) + "_" + path
}
