package boards2

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

func TestPostAddRepostTo(t *testing.T) {
	cases := []struct {
		name, title, body string
		dstBoard          *Board
		thread            *Post
		setup             func() *Post
		err               string
	}{
		{
			name:     "repost thread",
			title:    "Repost Title",
			body:     "Repost body",
			dstBoard: newBoard(42, "dst123", testutils.TestAddress("creatorDstBoard")),
			setup:    func() *Post { return createTestThread(t) },
		},
		{
			name:  "invalid repost from reply",
			setup: func() *Post { return createTestReply(t) },
			err:   "cannot repost non-thread post",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				repost  *Post
				creator = testutils.TestAddress("repostCreator")
				thread  = tc.setup()
			)

			createRepost := func() {
				repost = thread.AddRepostTo(creator, tc.title, tc.body, tc.dstBoard)
			}

			if tc.err != "" {
				uassert.PanicsWithMessage(t, tc.err, createRepost)
				return
			} else {
				uassert.NotPanics(t, createRepost)
			}

			r, found := tc.dstBoard.GetThread(repost.GetPostID())
			uassert.True(t, found)
			uassert.True(t, repost == r)
			uassert.Equal(t, tc.title, repost.GetTitle())
			uassert.Equal(t, tc.body, repost.GetBody())
			uassert.Equal(t, uint(thread.GetBoard().GetID()), uint(repost.GetRepostBoardID()))
		})
	}
}

func TestNewThread(t *testing.T) {
	creator := testutils.TestAddress("creator")
	member := testutils.TestAddress("member")
	title := "Test Title"
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
		"/r/demo/boards2$help&func=CreateReply&bid=%d&threadID=%d&postID=%d",
		uint(boardID),
		uint(threadID),
		uint(threadID),
	)
	repostURL := ufmt.Sprintf(
		"/r/demo/boards2$help&func=CreateRepost&bid=%d&postID=%d",
		uint(boardID),
		uint(threadID),
	)
	deleteURL := ufmt.Sprintf(
		"/r/demo/boards2$help&func=DeleteThread&bid=%d&threadID=%d",
		uint(boardID),
		uint(threadID),
	)

	thread := newPost(board, threadID, creator, title, body, threadID, 0, 0)

	uassert.True(t, thread.IsThread())
	uassert.Equal(t, uint(threadID), uint(thread.GetPostID()))
	uassert.False(t, thread.GetCreatedAt().IsZero())
	uassert.True(t, thread.GetUpdatedAt().IsZero())
	uassert.Equal(t, title, thread.GetTitle())
	uassert.Equal(t, body[:77]+"...", thread.GetSummary())
	uassert.Equal(t, url, thread.GetURL())
	uassert.Equal(t, replyURL, thread.GetReplyFormURL())
	uassert.Equal(t, repostURL, thread.GetRepostFormURL())
	uassert.Equal(t, deleteURL, thread.GetDeleteFormURL())
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

func TestThreadGetReply(t *testing.T) {
	cases := []struct {
		name   string
		thread *Post
		setup  func(thread *Post) (replyID PostID)
		found  bool
	}{
		{
			name:   "found",
			thread: createTestThread(t),
			setup: func(thread *Post) PostID {
				reply := thread.AddReply(testutils.TestAddress("replier"), "")
				return reply.GetPostID()
			},
			found: true,
		},
		{
			name:   "not found",
			thread: createTestThread(t),
			setup:  func(*Post) PostID { return 42 },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			replyID := tc.setup(tc.thread)

			reply, found := tc.thread.GetReply(replyID)

			uassert.Equal(t, tc.found, found)
			if reply != nil {
				uassert.Equal(t, uint(replyID), uint(reply.GetPostID()))
			}
		})
	}
}

func TestThreadDeleteReply(t *testing.T) {
	thread := createTestThread(t)
	cases := []struct {
		name  string
		setup func() PostID
		err   string
	}{
		{
			name: "ok",
			setup: func() PostID {
				reply := thread.AddReply(testutils.TestAddress("replier"), "")
				return reply.GetPostID()
			},
		},
		{
			name: "ok nested",
			setup: func() PostID {
				reply := thread.AddReply(testutils.TestAddress("replier"), "")
				return reply.AddReply(testutils.TestAddress("replier2"), "").GetPostID()
			},
		},
		{
			name:  "invalid",
			setup: func() PostID { return thread.GetPostID() },
			err:   "expected an ID of an inner reply",
		},
		{
			name:  "not found",
			setup: func() PostID { return 42 },
			err:   "reply not found in thread",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			replyID := tc.setup()

			err := thread.DeleteReply(replyID)

			if tc.err != "" {
				uassert.ErrorContains(t, err, tc.err)
				return
			}

			uassert.NoError(t, err)
			_, found := thread.GetReply(replyID)
			uassert.False(t, found)
		})
	}
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
		"/r/demo/boards2$help&func=CreateReply&bid=%d&threadID=%d&postID=%d",
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

func TestReplyGetReply(t *testing.T) {
	thread := createTestThread(t)
	parentReply := thread.AddReply(testutils.TestAddress("parentReplier"), "")
	cases := []struct {
		name  string
		setup func() PostID
		found bool
	}{
		{
			name: "found",
			setup: func() PostID {
				reply := parentReply.AddReply(testutils.TestAddress("replier"), "")
				return reply.GetPostID()
			},
			found: true,
		},
		{
			name:  "not found",
			setup: func() PostID { return 42 },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			replyID := tc.setup()

			reply, found := thread.GetReply(replyID)

			uassert.Equal(t, tc.found, found)
			if reply != nil {
				uassert.Equal(t, uint(replyID), uint(reply.GetPostID()))
			}
		})
	}
}

func TestReplyDeleteReply(t *testing.T) {
	thread := createTestThread(t)
	parentReply := thread.AddReply(testutils.TestAddress("replier"), "")
	reply := parentReply.AddReply(testutils.TestAddress("replier2"), "")

	// NOTE: Deleting a reply from a parent reply should eventually be suported
	uassert.PanicsWithMessage(t, "cannot delete reply from a non-thread post", func() {
		parentReply.DeleteReply(reply.GetPostID())
	})
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
