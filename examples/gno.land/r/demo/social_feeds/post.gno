package social_feeds

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
)

type PostID uint64

type Reaction struct {
	icon  string
	count uint64
}

var Categories []string = []string{
	"Reaction",
	"Comment",
	"Normal",
	"Article",
	"Picture",
	"Audio",
	"Video",
}

type Post struct {
	id        PostID
	parentID  PostID
	feedID    FeedID
	category  uint64
	metadata  string
	reactions avl.Tree
	creator   std.Address
	deleted   bool

	createdAt time.Time
	updatedAt time.Time
	deletedAt time.Time
}

func newPost(feed *Feed, id PostID, creator std.Address, parentID PostID, category uint64, metadata string) *Post {
	return &Post{
		id:        id,
		parentID:  parentID,
		feedID:    feed.id,
		category:  category,
		metadata:  metadata,
		reactions: avl.Tree{},
		creator:   creator,
		createdAt: time.Now(),
	}
}

func (post *Post) Update(category uint64, metadata string) {
	post.category = category
	post.metadata = metadata
	post.updatedAt = time.Now()
}

func (post *Post) Delete() {
	post.deleted = true
	post.deletedAt = time.Now()
}

// Always remove reaction if count = 0
func (post *Post) React(icon string, up bool) {
	count_, ok := post.reactions.Get(icon)
	count := 0

	if ok {
		count = count_.(int)
	}

	if up {
		count++
	} else {
		count--
	}

	if count <= 0 {
		post.reactions.Remove(icon)
	} else {
		post.reactions.Set(icon, count)
	}
}
