package boards2

import (
	"errors"
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/txlink"
)

const dateFormat = "2006-01-02 3:04pm MST"

type PostID uint64

func (id PostID) String() string {
	return strconv.Itoa(int(id))
}

func (id PostID) Key() string {
	return padZero(uint64(id), 10)
}

// A Post is a "thread" or a "reply" depending on context.
// A thread is a Post of a Board that holds other replies.
type Post struct {
	board         *Board
	id            PostID
	creator       std.Address
	title         string // optional
	body          string
	isHidden      bool
	replies       avl.Tree // Post.id -> *Post
	repliesAll    avl.Tree // Post.id -> *Post (all replies, for top-level posts)
	reposts       avl.Tree // Board.id -> Post.id
	flags         []Flag
	threadID      PostID  // original Post.id
	parentID      PostID  // parent Post.id (if reply or repost)
	repostBoardID BoardID // original Board.id (if repost)
	createdAt     time.Time
	updatedAt     time.Time
}

func newPost(board *Board, id PostID, creator std.Address, title, body string, threadID, parentID PostID, repostBoardID BoardID) *Post {
	return &Post{
		board:         board,
		id:            id,
		creator:       creator,
		title:         title,
		body:          body,
		replies:       avl.Tree{},
		repliesAll:    avl.Tree{},
		reposts:       avl.Tree{},
		threadID:      threadID,
		parentID:      parentID,
		repostBoardID: repostBoardID,
		createdAt:     time.Now(),
	}
}

func (post *Post) IsThread() bool {
	return post.parentID == 0
}

func (post *Post) GetBoard() *Board {
	return post.board
}

func (post *Post) GetPostID() PostID {
	return post.id
}

func (post *Post) GetParentID() PostID {
	return post.parentID
}

func (post *Post) GetRepostBoardID() BoardID {
	return post.repostBoardID
}

func (post *Post) GetCreator() std.Address {
	return post.creator
}

func (post *Post) GetTitle() string {
	return post.title
}

func (post *Post) GetBody() string {
	return post.body
}

func (post *Post) GetCreatedAt() time.Time {
	return post.createdAt
}

func (post *Post) GetUpdatedAt() time.Time {
	return post.updatedAt
}

func (post *Post) AddFlag(flag Flag) bool {
	// TODO: sort flags for fast search in case of big thresholds
	for _, v := range post.flags {
		if v.User == flag.User {
			return false
		}
	}

	post.flags = append(post.flags, flag)
	return true
}

func (post *Post) FlagsCount() int {
	return len(post.flags)
}

func (post *Post) SetVisible(isVisible bool) {
	post.isHidden = !isVisible
}

func (post *Post) Hidden() bool {
	return post.isHidden
}

func (post *Post) AddReply(creator std.Address, body string) *Post {
	board := post.board
	pid := board.incGetPostID()
	pKey := pid.Key()
	reply := newPost(board, pid, creator, "", body, post.threadID, post.id, 0)
	// TODO: Figure out how to remove this redundancy of data "replies==repliesAll" in threads
	post.replies.Set(pKey, reply)
	if post.threadID == post.id {
		post.repliesAll.Set(pKey, reply)
	} else {
		thread, _ := board.GetThread(post.threadID)
		thread.repliesAll.Set(pKey, reply)
	}
	return reply
}

func (post *Post) Update(title string, body string) {
	post.title = title
	post.body = body
	post.updatedAt = time.Now()
}

func (thread *Post) GetReply(pid PostID) (_ *Post, found bool) {
	v, found := thread.repliesAll.Get(pid.Key())
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

func (post *Post) AddRepostTo(creator std.Address, title, body string, dst *Board) *Post {
	if !post.IsThread() {
		panic("cannot repost non-thread post")
	}

	pid := dst.incGetPostID()
	repost := newPost(dst, pid, creator, title, body, pid, post.id, post.board.id)
	dst.threads.Set(pid.Key(), repost)
	if !dst.IsPrivate() {
		post.reposts.Set(dst.id.Key(), pid)
	}
	return repost
}

func (post *Post) DeleteReply(replyID PostID) error {
	if !post.IsThread() {
		// TODO: Allow removing replies from parent replies too
		panic("cannot delete reply from a non-thread post")
	}

	if post.id == replyID {
		return errors.New("expected an ID of an inner reply")
	}

	key := replyID.Key()
	v, removed := post.repliesAll.Remove(key)
	if !removed {
		return errors.New("reply not found in thread")
	}

	// TODO: Remove child replies too!
	reply := v.(*Post)
	if reply.parentID != post.id {
		parent, _ := post.GetReply(reply.parentID)
		parent.replies.Remove(key)
	} else {
		post.replies.Remove(key)
	}
	return nil
}

func (post *Post) GetSummary() string {
	return summaryOf(post.body, 80)
}

func (post *Post) GetURL() string {
	if post.IsThread() {
		return post.board.GetURLFromThreadID(post.id)
	}
	return post.board.GetURLFromReplyID(post.threadID, post.id)
}

func (post *Post) GetReplyFormURL() string {
	return txlink.Call("CreateReply",
		"bid", post.board.id.String(),
		"threadID", post.threadID.String(),
		"postID", post.id.String(),
	)
}

func (post *Post) GetRepostFormURL() string {
	return txlink.Call("CreateRepost",
		"bid", post.board.id.String(),
		"postID", post.id.String(),
	)
}

func (post *Post) GetDeleteFormURL() string {
	if post.IsThread() {
		return txlink.Call("DeleteThread",
			"bid", post.board.id.String(),
			"threadID", post.threadID.String(),
		)
	}
	return txlink.Call("DeleteReply",
		"bid", post.board.id.String(),
		"threadID", post.threadID.String(),
		"replyID", post.id.String(),
	)
}

func (post *Post) RenderSummary() string {
	if post.repostBoardID != 0 {
		dstBoard, found := getBoard(post.repostBoardID)
		if !found {
			panic("repost board does not exist")
		}

		thread, found := dstBoard.GetThread(PostID(post.parentID))
		if !found {
			return "reposted post does not exist"
		}
		return "Repost: " + post.GetSummary() + "\n" + thread.RenderSummary()
	}

	var (
		s       string
		postURL = post.GetURL()
	)

	if post.title != "" {
		s += "## " + newLink(summaryOf(post.title, 80), postURL) + "\n\n"
	}

	s += post.GetSummary() + "\n"
	s += "\\- " + newUserLink(post.creator) + ","
	s += " " + newLink(post.createdAt.Format(dateFormat), postURL)
	s += " \\[" + newLink("x", post.GetDeleteFormURL()) + "]"
	s += " (" + strconv.Itoa(post.replies.Size()) + " replies)"
	s += " (" + strconv.Itoa(post.reposts.Size()) + " reposts)" + "\n"
	return s
}

func (post *Post) Render(indent string, levels int) string {
	if post == nil {
		return "nil post"
	}

	var (
		s       string
		postURL = post.GetURL()
	)

	if post.title != "" {
		s += indent + "# " + post.title + "\n"
		s += indent + "\n"
	}

	s += indentBody(indent, post.body) + "\n" // TODO: indent body lines.
	s += indent + "\\- " + newUserLink(post.creator) + ", "
	s += newLink(post.createdAt.Format(dateFormat), postURL)
	s += " \\[" + newLink("reply", post.GetReplyFormURL()) + "]"
	if post.IsThread() {
		s += " \\[" + newLink("repost", post.GetRepostFormURL()) + "]"
	}
	s += " \\[" + newLink("x", post.GetDeleteFormURL()) + "]\n"

	if levels > 0 {
		if post.replies.Size() > 0 {
			post.replies.Iterate("", "", func(_ string, value interface{}) bool {
				s += indent + "\n"
				s += value.(*Post).Render(indent+"> ", levels-1)
				return false
			})
		}
	} else if post.replies.Size() > 0 {
		s += indent + "\n"
		s += indent + "_" + newLink("see all "+strconv.Itoa(post.replies.Size())+" replies", postURL) + "_\n"
	}
	return s
}

func (post *Post) RenderInner() string {
	if post.IsThread() {
		panic("unexpected thread")
	}

	var (
		parent    *Post
		parentID  = post.parentID
		threadID  = post.threadID
		thread, _ = post.board.GetThread(threadID) // TODO: This seems redundant (post == thread)
	)

	if thread.id == parentID {
		parent = thread
	} else {
		parent, _ = thread.GetReply(parentID)
	}

	s := "_" + newLink("see thread", post.board.GetURLFromThreadID(threadID)) + "_\n\n"
	s += parent.Render("", 0) + "\n"
	s += post.Render("> ", 5)
	return s
}
