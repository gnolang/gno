package social_feeds

import (
	"encoding/base64"
	"encoding/binary"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	ujson "gno.land/p/demo/ujson"
)

type PostID uint64

func (pid PostID) String() string {
	return strconv.Itoa(int(pid))
}

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
	id            PostID
	parentID      PostID
	feedID        FeedID
	category      uint64
	metadata      string
	reactions     avl.Tree // icon -> count
	comments      avl.Tree // Post.id -> *Post
	creator       std.Address
	tipAmount     uint64
	deleted       bool
	commentsCount uint64

	createdAt int64
	updatedAt int64
	deletedAt int64
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
		createdAt: time.Now().Unix(),
	}
}

func (post *Post) String() string {
	return post.ToJSON()
}

func (post *Post) Update(category uint64, metadata string) {
	post.category = category
	post.metadata = metadata
	post.updatedAt = time.Now().Unix()
}

func (post *Post) Delete() {
	post.deleted = true
	post.deletedAt = time.Now().Unix()
}

func (post *Post) Tip(from std.Address, to std.Address) {
	receivedCoins := std.GetOrigSend()
	amount := receivedCoins[0].Amount

	banker := std.GetBanker(std.BankerTypeOrigSend)
	// banker := std.GetBanker(std.BankerTypeRealmSend)
	coinsToSend := std.Coins{std.Coin{Denom: "ugnot", Amount: amount}}
	pkgaddr := std.GetOrigPkgAddr()

	banker.SendCoins(pkgaddr, to, coinsToSend)

	// Update tip amount
	post.tipAmount += uint64(amount)
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

func (post *Post) Render() string {
	return post.metadata
}

func (post *Post) FromJSON(jsonData string) {
	ast := ujson.TokenizeAndParse(jsonData)
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "id", CustomParser: func(node *ujson.JSONASTNode) {
			pid, _ := strconv.Atoi(node.Value)
			post.id = PostID(pid)
		}},
		{Key: "parentID", CustomParser: func(node *ujson.JSONASTNode) {
			pid, _ := strconv.Atoi(node.Value)
			post.parentID = PostID(pid)
		}},
		{Key: "feedID", CustomParser: func(node *ujson.JSONASTNode) {
			fid, _ := strconv.Atoi(node.Value)
			post.feedID = FeedID(fid)
		}},
		{Key: "category", Value: &post.category},
		{Key: "metadata", Value: &post.metadata},
		{Key: "reactions", CustomParser: func(node *ujson.JSONASTNode) {
			reactions := avl.NewTree()
			for _, child := range node.ObjectChildren {
				reactionCount := child.Value
				reactions.Set(child.Key, reactionCount)
			}
			post.reactions = *reactions
		}},
		{Key: "commentsCount", Value: &post.commentsCount},
		{Key: "creator", Value: &post.creator},
		{Key: "tipAmount", Value: &post.tipAmount},
		{Key: "deleted", Value: &post.deleted},
		{Key: "createdAt", Value: &post.createdAt},
		{Key: "updatedAt", Value: &post.updatedAt},
		{Key: "deletedAt", Value: &post.deletedAt},
	})
}

func (post *Post) ToJSON() string {
	reactionsKV := []ujson.FormatKV{}
	post.reactions.Iterate("", "", func(key string, value interface{}) bool {
		count := value.(int)
		data := ujson.FormatKV{Key: key, Value: count}
		reactionsKV = append(reactionsKV, data)
		return false
	})
	reactions := ujson.FormatObject(reactionsKV)

	postJSON := ujson.FormatObject([]ujson.FormatKV{
		{Key: "id", Value: uint64(post.id)},
		{Key: "parentID", Value: uint64(post.parentID)},
		{Key: "feedID", Value: uint64(post.feedID)},
		{Key: "category", Value: post.category},
		{Key: "metadata", Value: post.metadata},
		{Key: "reactions", Value: reactions, Raw: true},
		{Key: "creator", Value: post.creator},
		{Key: "tipAmount", Value: post.tipAmount},
		{Key: "deleted", Value: post.deleted},
		{Key: "commentsCount", Value: post.commentsCount},
		{Key: "createdAt", Value: post.createdAt},
		{Key: "updatedAt", Value: post.updatedAt},
		{Key: "deletedAt", Value: post.deletedAt},
	})
	return postJSON
}
