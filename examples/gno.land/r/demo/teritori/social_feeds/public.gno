package social_feeds

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/teritori/flags_index"
	"gno.land/p/demo/ufmt"
)

// Only registered user can create a new feed
// For the flexibility when testing, allow all user to create feed
func CreateFeed(name string) FeedID {
	pkgpath := std.CurrentRealmPath()

	fid := incGetFeedID()
	caller := std.PrevRealm().Addr()
	url := strings.Replace(pkgpath, "gno.land", "", -1) + ":" + name
	feed := newFeed(fid, url, name, caller)
	fidkey := feedIDKey(fid)
	gFeeds.Set(fidkey, feed)
	gFeedsByName.Set(name, feed)
	return feed.id
}

// Anyone can create a post in a existing feed, allow un-registered users also
func CreatePost(fid FeedID, parentID PostID, catetory uint64, metadata string) PostID {
	caller := std.PrevRealm().Addr()

	feed := mustGetFeed(fid)
	post := feed.AddPost(caller, parentID, catetory, metadata)
	return post.id
}

// Only post's owner can edit post
func EditPost(fid FeedID, pid PostID, category uint64, metadata string) {
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)
	post := feed.MustGetPost(pid)

	if caller != post.creator {
		panic("you are not creator of this post")
	}

	post.Update(category, metadata)
}

// Only feed creator/owner can call this
func SetOwner(fid FeedID, newOwner std.Address) {
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)

	if caller != feed.creator && caller != feed.owner {
		panic("you are not creator/owner of this feed")
	}

	feed.owner = newOwner
}

// Only feed creator/owner or post creator can delete the post
func DeletePost(fid FeedID, pid PostID) {
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)
	post := feed.MustGetPost(pid)

	if caller != post.creator && caller != feed.creator && caller != feed.owner {
		panic("you are nor creator of this post neither creator/owner of the feed")
	}

	post.Delete()

	// If post is comment then decrease comments count on parent
	if uint64(post.parentID) != 0 {
		parent := feed.MustGetPost(post.parentID)
		parent.commentsCount -= 1
	}
}

// Only feed owner can ban the post
func BanPost(fid FeedID, pid PostID, reason string) {
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)
	_ = feed.MustGetPost(pid)

	// For experimenting, we ban only the post for now
	// TODO: recursive delete/ban comments
	if caller != feed.owner {
		panic("you are owner of the feed")
	}

	feed.BanPost(pid)

	feed.flags.ClearFlagCount(getFlagID(fid, pid))
}

// Any one can react post
func ReactPost(fid FeedID, pid PostID, icon string, up bool) {
	feed := mustGetFeed(fid)
	post := feed.MustGetPost(pid)

	post.React(icon, up)
}

func TipPost(fid FeedID, pid PostID) {
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)
	post := feed.MustGetPost(pid)

	post.Tip(caller, post.creator)
}

// Get a list of flagged posts
// NOTE: We can support multi feeds in the future but for now we will have only 1 feed
// Return stringified list in format: postStr-count,postStr-count
func GetFlaggedPosts(fid FeedID, offset uint64, limit uint8) string {
	feed := mustGetFeed(fid)

	// Already sorted by count descending
	flags := feed.flags.GetFlags(uint64(limit), offset)

	var postList []string
	for _, flagCount := range flags {
		flagID := flagCount.FlagID

		feedID, postID := parseFlagID(flagID)
		if feedID != feed.id {
			continue
		}

		post := feed.GetPost(postID)
		postList = append(postList, ufmt.Sprintf("%s", post))
	}

	SEPARATOR := ","
	res := strings.Join(postList, SEPARATOR)
	return ufmt.Sprintf("[%s]", res)
}

// NOTE: due to bug of std.PrevRealm().Addr() return "" when query so we user this proxy function temporary
// in waiting of correct behaviour of std.PrevRealm().Addr()
func GetPosts(fid FeedID, parentID PostID, user string, categories []uint64, offset uint64, limit uint8) string {
	caller := std.PrevRealm().Addr()
	data := GetPostsWithCaller(fid, parentID, caller.String(), user, categories, offset, limit)
	return data
}

func GetPostsWithCaller(fid FeedID, parentID PostID, callerAddrStr string, user string, categories []uint64, offset uint64, limit uint8) string {
	// Return flagged posts, we process flagged posts differently using FlagIndex
	if len(categories) == 1 && categories[0] == uint64(9) {
		return GetFlaggedPosts(fid, offset, limit)
	}

	// BUG: normally std.PrevRealm().Addr() should return a value instead of empty
	// Fix is in progress on Gno side
	feed := mustGetFeed(fid)
	posts := getPosts(feed, parentID, callerAddrStr, user, categories, offset, limit)

	SEPARATOR := ","
	var postListStr []string

	for _, post := range posts {
		postListStr = append(postListStr, post.String())
	}

	res := strings.Join(postListStr, SEPARATOR)
	return ufmt.Sprintf("[%s]", res)
}

// user here is: filter by user
func getPosts(feed *Feed, parentID PostID, callerAddrStr string, user string, categories []uint64, offset uint64, limit uint8) []*Post {
	caller := std.Address(callerAddrStr)

	var posts []*Post
	var skipped uint64

	// Create an avlTree for optimizing the check
	requestedCategories := avl.NewTree()
	for _, category := range categories {
		catStr := strconv.FormatUint(category, 10)
		requestedCategories.Set(catStr, true)
	}

	feed.posts.ReverseIterate("", "", func(key string, value interface{}) bool {
		post := value.(*Post)

		postCatStr := strconv.FormatUint(post.category, 10)

		// NOTE: this search mechanism is not efficient, only for demo purpose
		if post.parentID == parentID && post.deleted == false {
			if requestedCategories.Size() > 0 && !requestedCategories.Has(postCatStr) {
				return false
			}

			if user != "" && std.Address(user) != post.creator {
				return false
			}

			// Filter hidden post
			flagID := getFlagID(feed.id, post.id)
			if feed.flags.HasFlagged(flagID, callerAddrStr) {
				return false
			}

			// Check if post is in hidden list
			value, exists := feed.hiddenPostsByUser.Get(caller.String())
			if exists {
				hiddenPosts := value.(*avl.Tree)
				// If post.id exists in hiddenPosts tree => that post is hidden
				if hiddenPosts.Has(post.id.String()) {
					return false
				}
			}

			if skipped < offset {
				skipped++
				return false
			}

			posts = append(posts, post)
		}

		if len(posts) == int(limit) {
			return true
		}

		return false
	})

	return posts
}

// Get comments list
func GetComments(fid FeedID, parentID PostID, offset uint64, limit uint8) string {
	return GetPosts(fid, parentID, "", []uint64{}, offset, limit)
}

// Get Post
func GetPost(fid FeedID, pid PostID) string {
	feed := mustGetFeed(fid)

	data, ok := feed.posts.Get(postIDKey(pid))
	if !ok {
		panic("Unable to get post")
	}

	post := data.(*Post)
	return post.String()
}

func FlagPost(fid FeedID, pid PostID) {
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)

	feed.FlagPost(caller, pid)
}

func HidePostForMe(fid FeedID, pid PostID) {
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)

	feed.HidePostForUser(caller, pid)
}

func UnHidePostForMe(fid FeedID, pid PostID) {
	caller := std.PrevRealm().Addr()
	feed := mustGetFeed(fid)

	feed.UnHidePostForUser(caller, pid)
}

func GetFlags(fid FeedID, limit uint64, offset uint64) string {
	feed := mustGetFeed(fid)

	type FlagCount struct {
		FlagID flags_index.FlagID
		Count  uint64
	}

	flags := feed.flags.GetFlags(limit, offset)

	var res []string
	for _, flag := range flags {
		res = append(res, ufmt.Sprintf("%s:%d", flag.FlagID, flag.Count))
	}

	return strings.Join(res, "|")
}

// TODO: allow only creator to call
func GetFeedByID(fid FeedID) *Feed {
	return mustGetFeed(fid)
}

// TODO: allow only admin to call
func ExportFeedData(fid FeedID) string {
	feed := mustGetFeed(fid)
	feedJSON := feed.ToJSON()
	return feedJSON
}

// TODO: allow only admin to call
func ImportFeedData(fid FeedID, jsonData string) {
	feed := mustGetFeed(fid)
	feed.FromJSON(jsonData)
}

// func MigrateFromPreviousFeed(fid feedsV7.FeedID) {
// 	// Get exported data from previous feeds
// 	jsonData := feedsV7.ExportFeedData(fid)
// 	ImportFeedData(FeedID(uint64(fid)), jsonData)
// }
