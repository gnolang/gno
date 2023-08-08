package social_feeds

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
)

type FeedID uint64

func (fid FeedID) String() string {
	return strconv.Itoa(int(fid))
}

type Feed struct {
	id        FeedID
	url       string
	name      string
	creator   std.Address
	owner     std.Address
	posts     avl.Tree // Post.id -> *Post
	createdAt time.Time
	deleted   avl.Tree

	postsCtr uint64
}

func newFeed(fid FeedID, url string, name string, creator std.Address) *Feed {
	if !reName.MatchString(name) {
		panic("invalid feed name: " + name)
	}

	if gFeedsByName.Has(name) {
		panic("feed already exists: " + name)
	}

	return &Feed{
		id:        fid,
		url:       url,
		name:      name,
		creator:   creator,
		owner:     creator,
		posts:     avl.Tree{},
		createdAt: time.Now(),
		deleted:   avl.Tree{},

		postsCtr: 0,
	}
}

func (feed *Feed) incGetPostID() PostID {
	feed.postsCtr++
	return PostID(feed.postsCtr)
}

func (feed *Feed) GetPost(pid PostID) *Post {
	pidkey := postIDKey(pid)
	post_, exists := feed.posts.Get(pidkey)
	if !exists {
		return nil
	}
	return post_.(*Post)
}

func (feed *Feed) MustGetPost(pid PostID) *Post {
	post := feed.GetPost(pid)
	if post == nil {
		panic("post does not exist")
	}
	return post
}

func (feed *Feed) AddPost(creator std.Address, parentID PostID, category uint64, metadata string) *Post {
	pid := feed.incGetPostID()
	pidkey := postIDKey(pid)

	post := newPost(feed, pid, creator, parentID, category, metadata)
	feed.posts.Set(pidkey, post)

	if parentID != 0 {
		parent := feed.MustGetPost(parentID)
		parent.comments.Set(pidkey, post)
	}
	return post
}

func (feed *Feed) Render() string {
	str := ""
	str += "There are " + intToString(feed.posts.Size()) + " post(s) \n\n"

	if feed.posts.Size() > 0 {
		feed.posts.Iterate("", "", func(key string, value interface{}) bool {
			if str != "" {
				str += "----------------------------------------\n"
			}

			post := value.(*Post)
			postUrl := "/r/demo/social_feeds_v4:" + feed.name + "/" + post.id.String()

			str += " * [PostID: " + post.id.String() +
				" - " + intToString(post.reactions.Size()) + " reactions " +
				" - " + intToString(post.comments.Size()) + " comments](" + postUrl + ")"
			return false
		})
	}
	return str
}
