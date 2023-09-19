package social_feeds

import (
	"encoding/binary"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/teritori/daodao/interfaces"
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

func (msg *ExecutableMessageBanPost) Type() string {
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

	feed := getFeed(msg.FeedID).(*Feed)
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

func (msg *ExecutableMessageBanPost) Binary() []byte {
	b := []byte{}

	t := msg.Type()
	b = binary.BigEndian.AppendUint16(b, uint16(len(t)))
	b = append(b, []byte(t)...)

	b = binary.BigEndian.AppendUint64(b, uint64(msg.FeedID))
	b = binary.BigEndian.AppendUint64(b, uint64(msg.PostID))

	b = binary.BigEndian.AppendUint16(b, uint16(len(msg.Reason)))
	b = append(b, []byte(msg.Reason)...)

	return b
}

func ExecutableMessageBanPostFromBinary(b []byte) *ExecutableMessageBanPost {
	msg := &ExecutableMessageBanPost{}

	if len(b) < 2 {
		panic("invalid length - less than 2")
	}
	tl := binary.BigEndian.Uint16(b[:2])
	b = b[2:]
	if len(b) < int(tl) {
		panic("invalid length - less than expected")
	}
	t := string(b[:tl])
	if t != msg.Type() {
		panic("invalid type")
	}
	b = b[tl:]

	if len(b) < 8 {
		panic("invalid length - less than 8")
	}
	msg.FeedID = FeedID(binary.BigEndian.Uint64(b[:8]))
	b = b[8:]

	if len(b) < 8 {
		panic("invalid length - less than 8")
	}
	msg.PostID = PostID(binary.BigEndian.Uint64(b[:8]))
	b = b[8:]

	rl := binary.BigEndian.Uint16(b[:2])
	b = b[2:]
	if len(b) < int(rl) {
		panic("invalid length - less than expected")
	}
	r := string(b[:rl])
	msg.Reason = r
	// b = b[rl:]

	return msg
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

func (h *BanPostHandler) Type() string {
	return ExecutableMessageBanPost{}.Type()
}

func (h *BanPostHandler) FromBinary(b []byte) dao_interfaces.ExecutableMessage {
	return ExecutableMessageBanPostFromBinary(b)
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
