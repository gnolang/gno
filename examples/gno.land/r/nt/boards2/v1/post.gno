package boards2

import (
	"errors"
	"std"
	"strconv"
	"strings"
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
	repostsCount  uint64
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

func (post *Post) IsRepost() bool {
	return post.repostBoardID != 0
}

func (post *Post) IsThread() bool {
	// repost threads also have parent ID
	return post.parentID == 0 || post.IsRepost()
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

func (post *Post) IsHidden() bool {
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

func (post *Post) HasReplies() bool {
	return post.replies.Size() > 0
}

func (thread *Post) GetReply(pid PostID) (_ *Post, found bool) {
	v, found := thread.repliesAll.Get(pid.Key())
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

func (post *Post) AddRepostTo(creator std.Address, repost *Post, dst *Board) {
	if !post.IsThread() {
		panic("cannot repost non-thread post")
	}

	post.repostsCount++
	dst.threads.Set(repost.id.Key(), repost)
	post.reposts.Set(dst.id.Key(), repost.id)
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

	// TODO: Shouldn't reply be hidden instead of deleted? Maybe replace reply by a deleted message.
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
	if post.IsThread() {
		return txlink.Call("CreateReply",
			"boardID", post.board.id.String(),
			"threadID", post.threadID.String(),
			"replyID", "0",
		)
	}
	return txlink.Call("CreateReply",
		"boardID", post.board.id.String(),
		"threadID", post.threadID.String(),
		"replyID", post.id.String(),
	)
}

func (post *Post) GetRepostFormURL() string {
	return txlink.Call("CreateRepost",
		"boardID", post.board.id.String(),
		"threadID", post.id.String(),
	)
}

func (post *Post) GetDeleteFormURL() string {
	if post.IsThread() {
		return txlink.Call("DeleteThread",
			"boardID", post.board.id.String(),
			"threadID", post.threadID.String(),
		)
	}
	return txlink.Call("DeleteReply",
		"boardID", post.board.id.String(),
		"threadID", post.threadID.String(),
		"replyID", post.id.String(),
	)
}

func (post *Post) GetEditFormURL() string {
	if post.IsThread() {
		return txlink.Call("EditThread",
			"boardID", post.board.id.String(),
			"threadID", post.threadID.String(),
			"title", post.GetTitle(),
			"body", post.GetBody(),
		)
	}

	return txlink.Call("EditReply",
		"boardID", post.board.id.String(),
		"threadID", post.threadID.String(),
		"replyID", post.id.String(),
		"body", post.GetBody(),
	)
}

func (post *Post) GetFlagFormURL() string {
	if post.IsThread() {
		return txlink.Call("FlagThread",
			"boardID", post.board.id.String(),
			"threadID", post.threadID.String(),
		)
	}

	return txlink.Call("FlagReply",
		"boardID", post.board.id.String(),
		"threadID", post.threadID.String(),
		"replyID", post.id.String(),
	)
}

func (post *Post) RenderSummary() string {
	var (
		b       strings.Builder
		postURL = post.GetURL()
		title   = post.title
	)

	b.WriteString("## " + newLink(summaryOf(title, 80), postURL) + "\n\n")
	b.WriteString("Created by " + newUserLink(post.creator) + " on " + post.createdAt.Format(dateFormat))
	b.WriteString(" - " + newHelpButtonLink("x", post.GetDeleteFormURL(), "delete thread") + "  \n")
	b.WriteString(strconv.Itoa(post.repliesAll.Size()) + " replies • ")
	b.WriteString(strconv.Itoa(post.reposts.Size()) + " reposts" + "\n")
	return b.String()
}

func (post *Post) renderSourcePost(indent string) (string, *Post) {
	if !post.IsRepost() {
		return "", nil
	}

	indent += "> "

	// TODO: figure out a way to decouple posts from a global storage.
	board, ok := getBoard(post.repostBoardID)
	if !ok {
		return indentBody(indent, "*⚠ Source board has been deleted*\n\n"), nil
	}

	srcPost, ok := board.GetThread(post.parentID)
	if !ok {
		return indentBody(indent, "*⚠ Source post has been deleted*\n\n"), nil
	}

	if srcPost.isHidden {
		return indentBody(indent, "*⚠ Source post has been flagged as inappropriate*\n\n"), nil
	}

	return indentBody(indent, srcPost.GetSummary()) + "\n\n", srcPost
}

// renderPostContent renders post text content (including repost body).
// Function will dump a predefined message instead of a body if post is hidden.
func (post *Post) renderPostContent(sb *strings.Builder, indent string) {
	if post.isHidden {
		// Flagged comment should be hidden, but replies still visible (see: #3480)
		// Flagged threads will be hidden by render function caller.
		sb.WriteString(indentBody(indent, "_⚠ Reply is hidden as it has been flagged as inappropriate_"))
		sb.WriteString("\n")
		return
	}

	srcContent, srcPost := post.renderSourcePost(indent)
	sb.WriteString(srcContent)
	sb.WriteString("\n")
	sb.WriteString(indentBody(indent, post.body))
	sb.WriteString("\n")

	if post.IsThread() {
		// Split content and controls for threads.
		sb.WriteString("\n")
	}

	// Buttons & counters
	sb.WriteString(indent)
	if !post.IsThread() {
		sb.WriteString("  \n" + indent)
	}

	sb.WriteString("Created by " + newUserLink(post.creator))
	sb.WriteString(" on " + post.createdAt.Format(dateFormat))

	if post.repostsCount > 0 {
		sb.WriteString(", ")
		sb.WriteString(strconv.FormatUint(post.repostsCount, 10))
		sb.WriteString(" repost(s)")
	}

	sb.WriteString(" - ")
	sb.WriteString(newButtonLink("reply", post.GetReplyFormURL()))

	if post.IsThread() {
		sb.WriteString(" ")
		sb.WriteString(newButtonLink("repost", post.GetRepostFormURL()))
	}

	sb.WriteString(" ")
	sb.WriteString(newButtonLink("edit", post.GetEditFormURL()))

	sb.WriteString(" ")
	sb.WriteString(newButtonLink("flag", post.GetFlagFormURL()))

	sb.WriteString(" ")
	sb.WriteString(newHelpButtonLink("x", post.GetDeleteFormURL(), "delete"))
	sb.WriteString("  \n")

	if post.IsRepost() && srcPost != nil {
		sb.WriteString("This thread is a repost of [another thread](" + srcPost.GetURL() + ")\n")
	}
}

func (post *Post) Render(p *PaginationOpts, indent string, levels int) string {
	if post == nil {
		return "nil post"
	}

	// TODO: pass a builder as arg into Render.
	var sb strings.Builder

	// Thread reposts might not have a title, if so get title from source thread
	title := post.title
	if post.IsRepost() && title == "" {
		if board, ok := getBoard(post.repostBoardID); ok {
			if src, ok := board.GetThread(post.parentID); ok {
				title = src.GetTitle()
			}
		}
	}

	sb.WriteString(indent)
	sb.WriteString("## " + title + "\n")
	sb.WriteString(indent + "\n")

	post.renderPostContent(&sb, indent)

	if post.replies.Size() == 0 {
		return sb.String()
	}

	if levels == 0 {
		sb.WriteString(indent + "\n")
		sb.WriteString(indent)
		sb.WriteString("_")
		sb.WriteString(newLink("see all "+strconv.Itoa(post.replies.Size())+" replies", post.GetURL()))
		sb.WriteString("_\n")
		return sb.String()
	}

	commentsIndent := indent + "> "
	page := p.Iterate(&post.replies, func(_ string, value interface{}) bool {
		reply := value.(*Post)

		sb.WriteString(indent)
		sb.WriteString("\n")
		sb.WriteString(reply.Render(nil, commentsIndent, levels-1))
		return false
	})

	if page != nil {
		sb.WriteString("\n---\n")
		sb.WriteString(page.Picker())
	}

	return sb.String()
}

func (post *Post) RenderInner() string {
	if post.IsThread() {
		panic("unexpected thread")
	}

	var (
		threadID  = post.threadID
		thread, _ = post.board.GetThread(threadID) // TODO: This seems redundant (post == thread)
	)

	s := "_" + newLink("see thread", post.board.GetURLFromThreadID(threadID)) + "_\n\n"

	// Fully render parent if it's not a repost.
	if !post.IsRepost() {
		var (
			parent   *Post
			parentID = post.parentID
		)

		if thread.id == parentID {
			parent = thread
		} else {
			parent, _ = thread.GetReply(parentID)
		}

		s += parent.Render(nil, "", 0) + "\n"
	}

	s += post.Render(nil, "> ", 5)
	return s
}
