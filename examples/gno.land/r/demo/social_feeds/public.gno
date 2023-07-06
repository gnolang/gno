package social_feeds

import (
	"encoding/base64"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
)

func CreateFeed(name string) FeedID {
	std.AssertOriginCall()
	fid := incGetFeedID()
	caller := std.PrevRealm().Addr()
	if usernameOf(caller) == "" {
		panic("only registered user can create feed")
	}
	url := "/r/demo/feeds:" + name
	feed := newFeed(fid, url, name, caller)
	fidkey := feedIDKey(fid)
	gFeeds.Set(fidkey, feed)
	gFeedsByName.Set(name, feed)
	return feed.id
}

// Anyone can create a post in a existing feed
// Anonymous user has to pay a configurable minimum fee
func CreatePost(fid FeedID, parentID PostID, catetory uint64, metadata string) PostID {
	std.AssertOriginCall()
	caller := std.PrevRealm().Addr()

	if usernameOf(caller) == "" {
		panic("please register user")
	}
	feed := mustGetFeed(fid)
	post := feed.AddPost(caller, parentID, catetory, metadata)
	return post.id
}

// Only post's owner can edit post
func EditPost(fid FeedID, pid PostID, category uint64, metadata string) {
	std.AssertOriginCall()
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)
	post := feed.MustGetPost(pid)

	if caller != post.creator {
		panic("you are not creator of this post")
	}

	post.Update(category, metadata)
}

// Only owner can delete the post
func DeletePost(fid FeedID, pid PostID) {
	std.AssertOriginCall()
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)
	post := feed.MustGetPost(pid)

	if caller != post.creator {
		panic("you are not creator of this post")
	}

	post.Delete()
}

// Any one can react post
func ReactPost(fid FeedID, pid PostID, icon string, up bool) {
	std.AssertOriginCall()
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)
	post := feed.MustGetPost(pid)

	post.React(icon, up)
}

// Get Posts list
func GetPosts(fid FeedID) string {
	std.AssertOriginCall()
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)

	var postList []string
	feed.posts.Iterate("", "", func(key string, value interface{}) bool {
		post := value.(*Post)
		postList = append(postList, post.String())
		return false
	})

	SEPARATOR := ","
	res := strings.Join(postList, SEPARATOR)
	return res
}

func checkAnonFee() bool {
	sent := std.GetOrigSend()
	anonFeeCoin := std.Coin{"ugnot", int64(gDefaultAnonFee)}
	if len(sent) == 1 && sent[0].IsGTE(anonFeeCoin) {
		return true
	}
	return false
}
