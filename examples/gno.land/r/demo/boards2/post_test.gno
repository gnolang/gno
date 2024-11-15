package boards

import (
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
)

func TestPostUpdate(t *testing.T) {
	board := newBoard(1, "test123", testutils.TestAddress("creator"))
	creator := testutils.TestAddress("creator")
	post := newPost(board, 1, creator, "Title", "Body", 1, 0, 0)
	title := "New Title"
	body := "New body"

	post.Update(title, body)

	uassert.Equal(t, title, post.GetTitle())
	uassert.Equal(t, body, post.GetBody())
	uassert.False(t, post.GetUpdatedAt().IsZero())
}

func TestNewThread(t *testing.T) {
	creator := testutils.TestAddress("creator")
	member := testutils.TestAddress("member")
	body := strings.Repeat("A", 82)
	boardID := BoardID(1)
	threadID := PostID(42)
	boardName := "test123"
	board := newBoard(boardID, boardName, creator)
	url := ufmt.Sprintf(
		"/r/demo/boards2:%s/%d",
		boardName,
		uint(threadID),
	)
	replyURL := ufmt.Sprintf(
		"/r/demo/boards2$help&func=CreateReply&bid=%d&threadid=%d&postid=%d",
		uint(boardID),
		uint(threadID),
		uint(threadID),
	)
	repostURL := ufmt.Sprintf(
		"/r/demo/boards2$help&func=CreateRepost&bid=%d&postid=%d",
		uint(boardID),
		uint(threadID),
	)
	deleteURL := ufmt.Sprintf(
		"/r/demo/boards2$help&func=DeleteThread&bid=%d&threadID=%d",
		uint(boardID),
		uint(threadID),
	)

	thread := newPost(board, threadID, creator, "Test", body, threadID, 0, 0)

	uassert.True(t, thread.IsThread())
	uassert.Equal(t, uint(threadID), uint(thread.GetPostID()))
	uassert.False(t, thread.GetCreatedAt().IsZero())
	uassert.True(t, thread.GetUpdatedAt().IsZero())
	uassert.Equal(t, body[:77]+"...", thread.GetSummary())
	uassert.Equal(t, url, thread.GetURL())
	uassert.Equal(t, replyURL, thread.GetReplyFormURL())
	uassert.Equal(t, repostURL, thread.GetRepostFormURL())
	uassert.Equal(t, deleteURL, thread.GetDeleteFormURL())
	uassert.True(t, thread.HasPermission(creator, PermissionEdit))
	uassert.True(t, thread.HasPermission(creator, PermissionDelete))
	uassert.False(t, thread.HasPermission(creator, Permission("unknown")))
	uassert.False(t, thread.HasPermission(member, PermissionEdit))
	uassert.False(t, thread.HasPermission(member, PermissionDelete))
	uassert.False(t, thread.HasPermission(member, Permission("unknown")))
}

func TestThreadAddReply(t *testing.T) {
	replier := testutils.TestAddress("replier")
	thread := createTestThread(t)
	threadID := uint(thread.GetPostID())
	body := "A reply"

	reply := thread.AddReply(replier, body)

	r, found := thread.GetReply(reply.GetPostID())
	uassert.True(t, found)
	uassert.True(t, reply == r)
	uassert.Equal(t, threadID+1, uint(reply.GetPostID()))
	uassert.Equal(t, reply.GetCreator(), replier)
	uassert.Equal(t, reply.GetBody(), body)
}

func TestThreadRenderSummary(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestThreadRender(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestThreadRenderInner(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestNewReply(t *testing.T) {
	creator := testutils.TestAddress("creator")
	member := testutils.TestAddress("member")
	body := strings.Repeat("A", 82)
	boardID := BoardID(1)
	threadID := PostID(42)
	parentID := PostID(1)
	replyID := PostID(2)
	boardName := "test123"
	board := newBoard(boardID, boardName, creator)
	url := ufmt.Sprintf(
		"/r/demo/boards2:%s/%d/%d",
		boardName,
		uint(threadID),
		uint(replyID),
	)
	replyURL := ufmt.Sprintf(
		"/r/demo/boards2$help&func=CreateReply&bid=%d&threadid=%d&postid=%d",
		uint(boardID),
		uint(threadID),
		uint(replyID),
	)
	deleteURL := ufmt.Sprintf(
		"/r/demo/boards2$help&func=DeleteReply&bid=%d&threadID=%d&replyID=%d",
		uint(boardID),
		uint(threadID),
		uint(replyID),
	)

	reply := newPost(board, replyID, creator, "", body, threadID, parentID, 0)

	uassert.False(t, reply.IsThread())
	uassert.Equal(t, uint(replyID), uint(reply.GetPostID()))
	uassert.False(t, reply.GetCreatedAt().IsZero())
	uassert.True(t, reply.GetUpdatedAt().IsZero())
	uassert.Equal(t, body[:77]+"...", reply.GetSummary())
	uassert.Equal(t, url, reply.GetURL())
	uassert.Equal(t, replyURL, reply.GetReplyFormURL())
	uassert.Equal(t, deleteURL, reply.GetDeleteFormURL())
	uassert.True(t, reply.HasPermission(creator, PermissionEdit))
	uassert.True(t, reply.HasPermission(creator, PermissionDelete))
	uassert.False(t, reply.HasPermission(creator, Permission("unknown")))
	uassert.False(t, reply.HasPermission(member, PermissionEdit))
	uassert.False(t, reply.HasPermission(member, PermissionDelete))
	uassert.False(t, reply.HasPermission(member, Permission("unknown")))
}

func TestReplyAddReply(t *testing.T) {
	replier := testutils.TestAddress("replier")
	thread := createTestThread(t)
	parentReply := thread.AddReply(testutils.TestAddress("parentReplier"), "")
	threadID := uint(thread.GetPostID())
	parentReplyID := uint(parentReply.GetPostID())
	body := "A child reply"

	reply := parentReply.AddReply(replier, body)

	r, found := thread.GetReply(reply.GetPostID())
	uassert.True(t, found)
	uassert.True(t, reply == r)
	uassert.Equal(t, parentReplyID, uint(reply.GetParentID()))
	uassert.Equal(t, parentReplyID+1, uint(reply.GetPostID()))
	uassert.Equal(t, reply.GetCreator(), replier)
	uassert.Equal(t, reply.GetBody(), body)
}

func TestReplyRender(t *testing.T) {
	t.Skip("TODO: implement")
}

func createTestThread(t *testing.T) *Post {
	t.Helper()

	creator := testutils.TestAddress("creator")
	board := newBoard(1, "test_board_123", creator)
	return board.AddThread(creator, "Title", "Body")
}

func createTestReply(t *testing.T) *Post {
	t.Helper()

	creator := testutils.TestAddress("replier")
	thread := createTestThread(t)
	return thread.AddReply(creator, "Test message")
}
