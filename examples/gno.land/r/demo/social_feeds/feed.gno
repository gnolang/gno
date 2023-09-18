package social_feeds

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/flags_index"
	"gno.land/p/demo/ufmt"
	ujson "gno.land/p/demo/ujson"
)

type FeedID uint64

func (fid FeedID) String() string {
	return strconv.Itoa(int(fid))
}

func (fid *FeedID) FromJSON(ast *ujson.JSONASTNode) {
	val, err := strconv.Atoi(ast.Value)
	if err != nil {
		panic(err)
	}
	*fid = FeedID(val)
}

func (fid FeedID) ToJSON() string {
	return strconv.Itoa(int(fid))
}

type Feed struct {
	id        FeedID
	url       string
	name      string
	creator   std.Address
	owner     std.Address
	posts     avl.Tree // pidkey -> *Post
	createdAt int64

	flags             *flags_index.FlagsIndex
	hiddenPostsByUser avl.Tree // std.Address => *avl.Tree (postID => bool)

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
		createdAt: time.Now().Unix(),
		flags:     flags_index.NewFlagsIndex(),
		postsCtr:  0,
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

	// If post is a comment then increase the comment count on parent
	if uint64(parentID) != 0 {
		parent := feed.MustGetPost(parentID)
		parent.commentsCount += 1
	}

	return post
}

func (feed *Feed) FlagPost(flagBy std.Address, pid PostID) {
	flagID := getFlagID(feed.id, pid)

	if feed.flags.HasFlagged(flagID, flagBy.String()) {
		panic("already flagged")
	}

	feed.flags.Flag(flagID, flagBy.String())
}

func (feed *Feed) BanPost(pid PostID) {
	pidkey := postIDKey(pid)
	_, removed := feed.posts.Remove(pidkey)
	if !removed {
		panic("post does not exist with id " + pid.String())
	}
}

func (feed *Feed) HidePostForUser(caller std.Address, pid PostID) {
	userAddr := caller.String()

	value, exists := feed.hiddenPostsByUser.Get(userAddr)
	var hiddenPosts *avl.Tree
	if exists {
		hiddenPosts = value.(*avl.Tree)
	} else {
		hiddenPosts = avl.NewTree()
		feed.hiddenPostsByUser.Set(userAddr, hiddenPosts)
	}

	if hiddenPosts.Has(pid.String()) {
		panic("PostID is already hidden: " + pid.String())
	}

	hiddenPosts.Set(pid.String(), true)
}

func (feed *Feed) UnHidePostForUser(userAddress std.Address, pid PostID) {
	value, exists := feed.hiddenPostsByUser.Get(userAddress.String())
	var hiddenPosts *avl.Tree
	if exists {
		hiddenPosts = value.(*avl.Tree)
		_, removed := hiddenPosts.Remove(pid.String())
		if !removed {
			panic("Post is not hidden: " + pid.String())
		}
	} else {
		panic("User has not hidden post: " + pid.String())
	}
}

func (feed *Feed) Render() string {
	pkgpath := std.CurrentRealmPath()

	str := ""
	str += ufmt.Sprintf("Feed: %s (ID: %s) - Owner: %s", feed.name, feed.id, feed.owner)
	str += "\n\n There are " + intToString(feed.posts.Size()) + " post(s) \n\n"

	if feed.posts.Size() > 0 {
		feed.posts.Iterate("", "", func(key string, value interface{}) bool {
			if str != "" {
				str += "\n"
			}

			post := value.(*Post)
			postUrl := strings.Replace(pkgpath, "gno.land", "", -1) + ":" + feed.name + "/" + post.id.String()

			str += " * [" +
				"PostID: " + post.id.String() +
				" - " + intToString(post.reactions.Size()) + " reactions " +
				" - " + ufmt.Sprintf("%d", post.tipAmount) + " tip amount" +
				"]" +
				"(" + postUrl + ")" +
				"\n"
			return false
		})

		str += "-------------------------\n"
		str += feed.flags.Dump()
	}

	str += "---------------------------------------\n"
	if feed.hiddenPostsByUser.Size() > 0 {
		str += "Hidden posts by users:\n\n"

		feed.hiddenPostsByUser.Iterate("", "", func(userAddr string, value interface{}) bool {
			hiddenPosts := value.(*avl.Tree)
			str += "\nUser address: " + userAddr + "\n"

			hiddenPosts.Iterate("", "", func(pid string, value interface{}) bool {
				str += "- PostID: " + pid + "\n"
				return false
			})

			return false
		})
	}

	return str
}

func (feed *Feed) ToJSON() string {
	posts := []ujson.FormatKV{}
	feed.posts.Iterate("", "", func(key string, value interface{}) bool {
		posts = append(posts, ujson.FormatKV{
			Key:   key,
			Value: value.(*Post),
		})
		return false
	})
	feedJSON := ujson.FormatObject([]ujson.FormatKV{
		{Key: "id", Value: uint64(feed.id)},
		{Key: "url", Value: feed.url},
		{Key: "name", Value: feed.name},
		{Key: "creator", Value: feed.creator},
		{Key: "owner", Value: feed.owner},
		{Key: "posts", Value: ujson.FormatObject(posts), Raw: true},
		{Key: "createdAt", Value: feed.createdAt},
		{Key: "postsCtr", Value: feed.postsCtr},
		// TODO: convert flags, hiddenPostsByUser
		// {Key: "flags", Value: feed.flags},
		// {Key: "hiddenPostsByUser", Value: feed.hiddenPostsByUser},
	})
	return feedJSON
}

func (feed *Feed) FromJSON(jsonData string) {
	ast := ujson.TokenizeAndParse(jsonData)
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "id", CustomParser: func(node *ujson.JSONASTNode) {
			fid, _ := strconv.Atoi(node.Value)
			feed.id = FeedID(fid)
		}},
		{Key: "url", Value: &feed.url},
		{Key: "name", Value: &feed.name},
		{Key: "creator", Value: &feed.creator},
		{Key: "owner", Value: &feed.owner},
		{Key: "posts", CustomParser: func(node *ujson.JSONASTNode) {
			posts := avl.NewTree()
			for _, child := range node.ObjectChildren {
				postNode := child.Value

				post := Post{}
				post.FromJSON(postNode.String())
				posts.Set(child.Key, &post)
			}
			feed.posts = *posts
		}},
		{Key: "createdAt", Value: &feed.createdAt},
		{Key: "postsCtr", Value: &feed.postsCtr},
	})
}
