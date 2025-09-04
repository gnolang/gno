package boards2

import (
	"chain/runtime"
	"strings"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
)

func TestPostFlag(t *testing.T) {
	addr := testutils.TestAddress("creator")
	post := createTestThread(t)

	uassert.True(t, post.Flag(addr, "foo"))
	uassert.False(t, post.Flag(addr, "foo"), "should reject flag from duplicate user")
	uassert.Equal(t, post.FlagsCount(), 1)
}

func TestPostRepost(t *testing.T) {
	// TODO: Improve this unit test
	addr := testutils.TestAddress("creatorDstBoard")
	perms := createBasicBoardPermissions(addr)
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
			dstBoard: newBoard(42, "dst123", addr, perms),
			setup:    func() *Post { return createTestThread(t) },
		},
		{
			name:  "invalid repost from reply",
			setup: func() *Post { return createTestReply(t) },
			err:   "post must be a thread to be reposted to another board",
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
				repost = thread.Repost(creator, tc.dstBoard, tc.title, tc.body)
			}

			if tc.err != "" {
				uassert.PanicsWithMessage(t, tc.err, createRepost)
				return
			} else {
				uassert.NotPanics(t, createRepost)
			}

			r, found := tc.dstBoard.GetThread(repost.ID)
			uassert.True(t, found)
			uassert.True(t, repost == r)
			uassert.Equal(t, tc.title, repost.Title)
			uassert.Equal(t, tc.body, repost.Body)
			uassert.Equal(t, uint(thread.Board.ID), uint(repost.RepostBoardID))
		})
	}
}

func TestNewThread(t *testing.T) {
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/gnoland/boards2/v1"))

	creator := testutils.TestAddress("creator")
	title := "Test Title"
	body := strings.Repeat("A", 82)
	boardID := BoardID(1)
	threadID := PostID(42)
	boardName := "test123"
	perms := createBasicBoardPermissions(creator)
	board := newBoard(boardID, boardName, creator, perms)
	url := ufmt.Sprintf(
		"/r/gnoland/boards2/v1:%s/%d",
		boardName,
		uint(threadID),
	)
	replyURL := ufmt.Sprintf(
		"/r/gnoland/boards2/v1$help&func=CreateReply&boardID=%d&body=&replyID=0&threadID=%d",
		uint(boardID),
		uint(threadID),
	)
	editURL := ufmt.Sprintf(
		"/r/gnoland/boards2/v1$help&func=EditThread&boardID=%d&body=%s&threadID=%d&title=%s",
		uint(boardID),
		body,
		uint(threadID),
		strings.ReplaceAll(title, " ", "+"),
	)
	repostURL := ufmt.Sprintf(
		"/r/gnoland/boards2/v1$help&func=CreateRepost&boardID=%d&body=&destinationBoardID=&threadID=%d&title=",
		uint(boardID),
		uint(threadID),
	)
	deleteURL := ufmt.Sprintf(
		"/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=%d&threadID=%d",
		uint(boardID),
		uint(threadID),
	)
	flagURL := ufmt.Sprintf(
		"/r/gnoland/boards2/v1$help&func=FlagThread&boardID=%d&reason=&threadID=%d",
		uint(boardID),
		uint(threadID),
	)

	thread := newPost(board, threadID, threadID, creator, title, body)

	uassert.True(t, thread.IsThread())
	uassert.Equal(t, uint(threadID), uint(thread.ID))
	uassert.False(t, thread.CreatedAt().IsZero())
	uassert.True(t, thread.UpdatedAt.IsZero())
	uassert.Equal(t, title, thread.Title)
	uassert.Equal(t, body[:77]+"...", thread.Summary())
	uassert.False(t, thread.HasReplies())
	uassert.Equal(t, url, makeThreadURI(thread))
	uassert.Equal(t, replyURL, makeCreateReplyURI(thread))
	uassert.Equal(t, editURL, makeEditPostURI(thread))
	uassert.Equal(t, repostURL, makeCreateRepostURI(thread))
	uassert.Equal(t, deleteURL, makeDeletePostURI(thread))
	uassert.Equal(t, flagURL, makeFlagURI(thread))
}

func TestThreadAddReply(t *testing.T) {
	replier := testutils.TestAddress("replier")
	thread := createTestThread(t)
	threadID := uint(thread.ID)
	body := "A reply"

	reply := thread.AddReply(replier, body)

	r, found := thread.GetReply(reply.ID)
	uassert.True(t, found)
	uassert.True(t, reply == r)
	uassert.Equal(t, threadID+1, uint(reply.ID))
	uassert.Equal(t, reply.Creator, replier)
	uassert.Equal(t, reply.Body, body)
	uassert.True(t, thread.HasReplies())
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
				return reply.ID
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
				uassert.Equal(t, uint(replyID), uint(reply.ID))
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
				return reply.ID
			},
		},
		{
			name: "ok nested",
			setup: func() PostID {
				reply := thread.AddReply(testutils.TestAddress("replier"), "")
				return reply.AddReply(testutils.TestAddress("replier2"), "").ID
			},
		},
		{
			name:  "invalid",
			setup: func() PostID { return thread.ID },
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
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/gnoland/boards2/v1"))

	creator := testutils.TestAddress("creator")
	body := strings.Repeat("A", 82)
	boardID := BoardID(1)
	threadID := PostID(42)
	parentID := PostID(1)
	replyID := PostID(2)
	boardName := "test123"
	perms := createBasicBoardPermissions(creator)
	board := newBoard(boardID, boardName, creator, perms)
	url := ufmt.Sprintf(
		"/r/gnoland/boards2/v1:%s/%d/%d",
		boardName,
		uint(threadID),
		uint(replyID),
	)
	replyURL := ufmt.Sprintf(
		"/r/gnoland/boards2/v1$help&func=CreateReply&boardID=%d&body=&replyID=%d&threadID=%d",
		uint(boardID),
		uint(replyID),
		uint(threadID),
	)
	deleteURL := ufmt.Sprintf(
		"/r/gnoland/boards2/v1$help&func=DeleteReply&boardID=%d&replyID=%d&threadID=%d",
		uint(boardID),
		uint(replyID),
		uint(threadID),
	)

	reply := newPost(board, threadID, replyID, creator, "", body)
	reply.ParentID = parentID

	uassert.False(t, reply.IsThread())
	uassert.Equal(t, uint(replyID), uint(reply.ID))
	uassert.False(t, reply.CreatedAt().IsZero())
	uassert.True(t, reply.UpdatedAt.IsZero())
	uassert.False(t, reply.HasReplies())
	uassert.Equal(t, body[:77]+"...", reply.Summary())
	uassert.Equal(t, url, makeReplyURI(reply))
	uassert.Equal(t, replyURL, makeCreateReplyURI(reply))
	uassert.Equal(t, deleteURL, makeDeletePostURI(reply))
}

func TestReplyAddReply(t *testing.T) {
	replier := testutils.TestAddress("replier")
	thread := createTestThread(t)
	parentReply := thread.AddReply(testutils.TestAddress("parentReplier"), "")
	parentReplyID := uint(parentReply.ID)
	body := "A child reply"

	reply := parentReply.AddReply(replier, body)

	r, found := thread.GetReply(reply.ID)
	uassert.True(t, found)
	uassert.True(t, reply == r)
	uassert.Equal(t, parentReplyID, uint(reply.ParentID))
	uassert.Equal(t, parentReplyID+1, uint(reply.ID))
	uassert.Equal(t, reply.Creator, replier)
	uassert.Equal(t, reply.Body, body)
	uassert.False(t, reply.HasReplies())
	uassert.True(t, parentReply.HasReplies())
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
				return reply.ID
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
				uassert.Equal(t, uint(replyID), uint(reply.ID))
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
		parentReply.DeleteReply(reply.ID)
	})
}

func TestReplyRender(t *testing.T) {
	t.Skip("TODO: implement")
}

func createTestThread(t *testing.T) *Post {
	t.Helper()

	creator := testutils.TestAddress("creator")
	perms := createBasicBoardPermissions(creator)
	board := newBoard(1, "test_board_123", creator, perms)
	return board.AddThread(creator, "Title", "Body")
}

func createTestReply(t *testing.T) *Post {
	t.Helper()

	creator := testutils.TestAddress("replier")
	thread := createTestThread(t)
	return thread.AddReply(creator, "Test message")
}
