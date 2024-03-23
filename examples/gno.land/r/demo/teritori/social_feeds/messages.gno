package social_feeds

import (
	"strings"

	"gno.land/p/demo/teritori/dao_interfaces"
	"gno.land/p/demo/teritori/ujson"
)

var PKG_PATH = "gno.land/r/demo/teritori/social_feeds"

// Ban a post
type ExecutableMessageBanPost struct {
	dao_interfaces.ExecutableMessage

	FeedID FeedID
	PostID PostID
	Reason string
}

func (msg ExecutableMessageBanPost) Type() string {
	return "gno.land/r/demo/teritori/social_feeds.BanPost"
}

func (msg *ExecutableMessageBanPost) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "feedId", Value: msg.FeedID},
		{Key: "postId", Value: msg.PostID},
		{Key: "reason", Value: msg.Reason},
	})
}

func (msg *ExecutableMessageBanPost) String() string {
	var ss []string
	ss = append(ss, msg.Type())

	feed := getFeed(msg.FeedID)
	s := ""

	if feed != nil {
		s += "Feed: " + feed.name + " (" + feed.id.String() + ")"

		post := feed.GetPost(msg.PostID)
		if post != nil {
			s += "\n Post: " + post.id.String()
		} else {
			s += "\n Post: " + msg.PostID.String() + " (not found)"
		}
	} else {
		s += "Feed: " + msg.FeedID.String() + " (not found)"
	}

	s += "\nReason: " + msg.Reason

	ss = append(ss, s)

	return strings.Join(ss, "\n---\n")
}

type BanPostHandler struct {
	dao_interfaces.MessageHandler
}

func NewBanPostHandler() *BanPostHandler {
	return &BanPostHandler{}
}

func (h *BanPostHandler) Execute(iMsg dao_interfaces.ExecutableMessage) {
	msg := iMsg.(*ExecutableMessageBanPost)
	BanPost(msg.FeedID, msg.PostID, msg.Reason)
}

func (h BanPostHandler) Type() string {
	return ExecutableMessageBanPost{}.Type()
}

func (h *BanPostHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	msg := &ExecutableMessageBanPost{}
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "feedId", Value: &msg.FeedID},
		{Key: "postId", Value: &msg.PostID},
		{Key: "reason", Value: &msg.Reason},
	})
	return msg
}
