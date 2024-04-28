package boards

import (
	"strconv"

	"gno.land/p/demo/boards"
)

func GetReplyFormURL(post *boards.Post) string {
	board := post.Board()
	bid := board.Id()
	return "/r/demo/boards?help&__func=CreateReply" +
		"&bid=" + bid.String() +
		"&threadid=" + post.ThreadID().String() +
		"&postid=" + post.Id().String() +
		"&body.type=textarea"
}

func GetRepostFormURL(post *boards.Post) string {
	board := post.Board()
	bid := board.Id()
	return "/r/demo/boards?help&__func=CreateRepost" +
		"&bid=" + bid.String() +
		"&postid=" + post.Id().String() +
		"&title.type=textarea" +
		"&body.type=textarea" +
		"&dstBoardID.type=textarea"
}

func GetDeleteFormURL(post *boards.Post) string {
	board := post.Board()
	bid := board.Id()
	return "/r/demo/boards?help&__func=DeletePost" +
		"&bid=" + bid.String() +
		"&threadid=" + post.ThreadID().String() +
		"&postid=" + post.Id().String()
}

func RenderSummary(post *boards.Post) string {
	if post.RepostBoard() != 0 {
		dstBoard := getBoard(post.RepostBoard())
		if dstBoard == nil {
			panic("repostBoard does not exist")
		}
		thread := dstBoard.GetThread(boards.PostID(post.ParentID()))
		if thread == nil {
			return "reposted post does not exist"
		}
		return "Repost: " + post.GetSummary() + "\n" + RenderSummary(thread)
	}

	replies := post.Replies()
	reposts := post.Reposts()

	str := ""
	if post.Title() != "" {
		str += "## [" + summaryOf(post.Title(), 80) + "](" + post.GetURL() + ")\n"
		str += "\n"
	}
	str += post.GetSummary() + "\n"
	str += "\\- " + displayAddressMD(post.Creator()) + ","
	str += " [" + post.CreatedAt().Format("2006-01-02 3:04pm MST") + "](" + post.GetURL() + ")"
	str += " \\[[x](" + GetDeleteFormURL(post) + ")]"
	str += " (" + strconv.Itoa(replies.Size()) + " replies)"
	str += " (" + strconv.Itoa(reposts.Size()) + " reposts)" + "\n"
	return str
}

func RenderPost(post *boards.Post, indent string, levels int) string {
	if post == nil {
		return "nil post"
	}
	str := ""
	if post.Title() != "" {
		str += indent + "# " + post.Title() + "\n"
		str += indent + "\n"
	}
	str += indentBody(indent, post.Body()) + "\n" // TODO: indent body lines.
	str += indent + "\\- " + displayAddressMD(post.Creator()) + ", "
	str += "[" + post.CreatedAt().Format("2006-01-02 3:04pm (MST)") + "](" + post.GetURL() + ")"
	str += " \\[[reply](" + GetReplyFormURL(post) + ")]"
	if post.IsThread() {
		str += " \\[[repost](" + GetRepostFormURL(post) + ")]"
	}
	str += " \\[[x](" + GetDeleteFormURL(post) + ")]\n"

	replies := post.Replies()
	if levels > 0 {
		if replies.Size() > 0 {
			replies.Iterate("", "", func(key string, value interface{}) bool {
				str += indent + "\n"
				str += RenderPost(value.(*boards.Post), indent+"> ", levels-1)
				return false
			})
		}
	} else {
		if replies.Size() > 0 {
			str += indent + "\n"
			str += indent + "_[see all " + strconv.Itoa(replies.Size()) + " replies](" + post.GetURL() + ")_\n"
		}
	}
	return str
}

// render reply and link to context thread
func RenderInner(post *boards.Post) string {
	if post.IsThread() {
		panic("unexpected thread")
	}
	board := post.Board()
	threadID := post.ThreadID()
	parentID := post.ParentID()
	str := ""
	str += "_[see thread](" + board.GetURLFromThreadAndReplyID(
		threadID, 0) + ")_\n\n"

	thread := board.GetThread(post.ThreadID())
	var parent *boards.Post
	if thread.Id() == parentID {
		parent = thread
	} else {
		parent = thread.GetReply(parentID)
	}
	str += RenderPost(parent, "", 0)
	str += "\n"
	str += RenderPost(post, "> ", 5)
	return str
}
