package boardsv2

import (
	"strconv"
)

// TODO: Locking
// - Locking a board means, you can not create new threads, and you can not comment in existing ones
// - Locking a thread means, you can not comment in this thread anymore

// TODO: Move boards (or App) to `boardsv2`
type Boards struct {
	// NOTE: We might want different AVL trees to avoid using level prefixes
	posts   avl.Tree // string(Post.Level + Post.CreatedAt + slug) -> *Post (post, comment, poll)
	locking lockingplugin.Plugin
}

// TODO: Support pagination Start/End (see pager implementation)
func (b Boards) Iterate(level int, path string, fn func(*Post) bool) bool        {}
func (b Boards) ReverseIterate(level int, path string, fn func(*Post) bool) bool {}

func (b *Boards) Lock(path string) {
	post := b.Get(LevelBoard, path) // Otherwise we try LevelPost
	if err := b.locking.Lock(post); err != nil {
		panic(err)
	}
}

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
//
// LEVEL_BOARD/POST/POST-2/COMMENT/COMMENT-2 (deprecated)
// LEVEL_TIMESTAMP_BOARD/POST/COMMENT
//
// :board/post/comment

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
	// TODO: Add timestamp to key
	return strconv.Itoa(level) + "_" + path
}
