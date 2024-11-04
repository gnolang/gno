package post

import "time"

/*
alicePost = &Post { content: "foo" } (0x001)
bobFork := &Post { Origial: alicePost (0x001) }

//1. Check gc behavior in realm for forks

---
alicePost := &(*alicePost) (0x002)
alicePost.content = "new content"

bobFork := &Post { Origial: uintptr(0x001) }
---
type Post struct {
	ID int
	Level int
}

package reddit

// explore with plugins
// - boardsv2
// 	- pkg/general
//  - pkg/reddit
var (
	rating avl.Tree
)

genericPost := Post{}
reddit.UpvotePost(genericPost.ID)
*/

// Blog example
// Home
// - post 1 (content: title, body, author, label, timestamp)
//   - post 1.1 (body, author) (thread)
//   - post 1.1.1 (comment to a thread but also a new thread)
//   - post 1.1.1.1
//   - post 1.2 (thread)
//
// - post 2
// - post 3
//
// Reddit example
// Home
// - post 1 (title, body) (board)
//   - post 1.1 (title, body) (sub-board)
//   - post 1.1.1 (title, body, label)
//   - post 1.1.1.1 (comment, thread)
type Post struct {
	ID        string
	Content   Content // title, body, label, author, other metadata...
	Level     int
	Base      *Post
	Children  []*Post
	Forks     []*Post
	UpdatedAt time.Time
	CreatedAt time.Time // the time when created by user or forked.
	Creator   std.Address
}

// create plugins for Post type <
// upvoting < implement first plugin
// define public API for plugin, post packages and boardsv2
// moderation
//
// plugin ideas:
// - visibility
// - upcoting
// - acess control > you shouldn't be able to answer to the boards yo're not invited
// - moedaration (ban certain posts -this could be through a dao in the future)

func New(s Storage) Post {

}

func Create(c Content) *Post {

}

func (p *Post) NextIncrementalKey(base string) string {

}

// func (p *Post) Append() error {
//
// }
