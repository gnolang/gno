package boards2

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
)

const dateFormat = "2006-01-02 3:04pm MST"

// PostID defines a type for Post (Threads/Replies) identifiers.
type PostID uint64

// String returns the ID as a string.
func (id PostID) String() string {
	return strconv.Itoa(int(id))
}

// Key returns the ID as a string with 10 characters padded with zeroes.
// This value can be used for indexing by ID.
func (id PostID) Key() string {
	return padZero(uint64(id), 10)
}

// A Post is a "thread" or a "reply" depending on context.
// A thread is a Post of a Board that holds other replies.
type Post struct {
	ID            PostID
	Board         *Board
	Creator       address
	Title         string
	Body          string
	Hidden        bool
	Readonly      bool
	ThreadID      PostID  // Original Post.ID
	ParentID      PostID  // Parent Post.ID (if reply or repost)
	RepostBoardID BoardID // Original Board.ID (if repost)
	UpdatedAt     time.Time

	flags      avl.Tree // std.Address -> string(reason)
	replies    avl.Tree // Post.ID -> *Post
	repliesAll avl.Tree // Post.ID -> *Post (all replies, for top-level posts)
	reposts    avl.Tree // Board.ID -> Post.ID
	createdAt  time.Time
}

func newPost(board *Board, threadID, id PostID, creator address, title, body string) *Post {
	return &Post{
		Board:     board,
		ThreadID:  threadID,
		ID:        id,
		Creator:   creator,
		Title:     title,
		Body:      body,
		createdAt: time.Now(),
	}
}

// CreatedAt returns the time when post was created.
func (post *Post) CreatedAt() time.Time {
	return post.createdAt
}

// IsRepost checks if current post is repost.
func (post *Post) IsRepost() bool {
	return post.RepostBoardID != 0
}

// IsThread checks if current post is a thread.
func (post *Post) IsThread() bool {
	// repost threads also have parent ID
	return post.ParentID == 0 || post.IsRepost()
}

// Flag add a flag to the post.
// It returns false when the user flagging the post already flagged it.
func (post *Post) Flag(user address, reason string) bool {
	if post.flags.Has(user.String()) {
		return false
	}

	post.flags.Set(user.String(), reason)
	return true
}

// FlagsCount returns the number of time post was flagged.
func (post *Post) FlagsCount() int {
	return post.flags.Size()
}

// AddReply adds a new reply to the post.
// Replies can be added to threads and also to other replies.
func (post *Post) AddReply(creator address, body string) *Post {
	board := post.Board
	pid := board.generateNextPostID()
	pKey := pid.Key()
	reply := newPost(board, post.ThreadID, pid, creator, "", body)
	reply.ParentID = post.ID
	// TODO: Figure out how to remove this redundancy of data "replies==repliesAll" in threads
	post.replies.Set(pKey, reply)
	if post.ThreadID == post.ID {
		post.repliesAll.Set(pKey, reply)
	} else {
		thread, _ := board.GetThread(post.ThreadID)
		thread.repliesAll.Set(pKey, reply)
	}
	return reply
}

// HasReplies checks if post has replies.
func (post *Post) HasReplies() bool {
	return post.replies.Size() > 0
}

// Get returns a post reply.
func (thread *Post) GetReply(pid PostID) (_ *Post, found bool) {
	v, found := thread.repliesAll.Get(pid.Key())
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

// Repost reposts a thread into another boards.
func (post *Post) Repost(creator address, dst *Board, title, body string) *Post {
	if !post.IsThread() {
		panic("post must be a thread to be reposted to another board")
	}

	repost := dst.AddThread(creator, title, body)
	repost.ParentID = post.ID
	repost.RepostBoardID = post.Board.ID

	dst.threads.Set(repost.ID.Key(), repost)
	post.reposts.Set(dst.ID.Key(), repost.ID)
	return repost
}

// DeleteReply deletes a reply from a thread.
func (post *Post) DeleteReply(replyID PostID) error {
	if !post.IsThread() {
		// TODO: Allow removing replies from parent replies too
		panic("cannot delete reply from a non-thread post")
	}

	if post.ID == replyID {
		return errors.New("expected an ID of an inner reply")
	}

	key := replyID.Key()
	v, removed := post.repliesAll.Remove(key)
	if !removed {
		return errors.New("reply not found in thread")
	}

	// TODO: Shouldn't reply be hidden instead of deleted? Maybe replace reply by a deleted message.
	reply := v.(*Post)
	if reply.ParentID != post.ID {
		parent, _ := post.GetReply(reply.ParentID)
		parent.replies.Remove(key)
	} else {
		post.replies.Remove(key)
	}
	return nil
}

// Summary return a summary of the post's body.
// It returns the body making sure that the length is limited to 80 characters.
func (post *Post) Summary() string {
	return summaryOf(post.Body, 80)
}

func (post *Post) RenderSummary() string {
	var (
		b             strings.Builder
		postURI       = makeThreadURI(post)
		threadSummary = summaryOf(post.Title, 80)
		creatorLink   = md.UserLink(post.Creator.String())
		date          = post.CreatedAt().Format(dateFormat)
	)

	b.WriteString(md.Bold("≡ "+md.Link(threadSummary, postURI)) + "  \n")
	b.WriteString("Created by " + creatorLink + " on " + date + "  \n")

	status := []string{
		strconv.Itoa(post.repliesAll.Size()) + " replies",
		strconv.Itoa(post.reposts.Size()) + " reposts",
	}
	b.WriteString(md.Bold(strings.Join(status, " • ")) + "\n")
	return b.String()
}

func (post *Post) renderSourcePost(indent string) (string, *Post) {
	if !post.IsRepost() {
		return "", nil
	}

	indent += "> "

	// TODO: figure out a way to decouple posts from a global storage.
	board, ok := getBoard(post.RepostBoardID)
	if !ok {
		// TODO: Boards can't be deleted so this might be redundant
		return indentBody(indent, md.Italic("⚠ Source board has been deleted")+"\n"), nil
	}

	srcPost, ok := board.GetThread(post.ParentID)
	if !ok {
		return indentBody(indent, md.Italic("⚠ Source post has been deleted")+"\n"), nil
	}

	if srcPost.Hidden {
		return indentBody(indent, md.Italic("⚠ Source post has been flagged as inappropriate")+"\n"), nil
	}

	return indentBody(indent, srcPost.Summary()) + "\n\n", srcPost
}

// renderPostContent renders post text content (including repost body).
// Function will dump a predefined message instead of a body if post is hidden.
func (post *Post) renderPostContent(sb *strings.Builder, indent string, levels int) {
	if post.Hidden {
		// Flagged comment should be hidden, but replies still visible (see: #3480)
		// Flagged threads will be hidden by render function caller.
		sb.WriteString(indentBody(indent, md.Italic("⚠ Reply is hidden as it has been flagged as inappropriate")) + "\n")
		return
	}

	srcContent, srcPost := post.renderSourcePost(indent)
	if post.IsRepost() && srcPost != nil {
		originLink := md.Link("another thread", makeThreadURI(srcPost))
		sb.WriteString("  \nThis thread is a repost of " + originLink + ": \n")
	}

	sb.WriteString(srcContent)

	if post.IsRepost() && srcPost == nil && len(post.Body) > 0 {
		// Add a newline to separate source deleted message from repost body content
		sb.WriteString("\n")
	}

	sb.WriteString(indentBody(indent, post.Body))
	sb.WriteString("\n")

	if post.IsThread() {
		// Split content and controls for threads.
		sb.WriteString("\n")
	}

	// Buttons & counters
	sb.WriteString(indent)
	if !post.IsThread() {
		sb.WriteString("  \n")
		sb.WriteString(indent)
	}

	creatorLink := md.UserLink(post.Creator.String())
	date := post.CreatedAt().Format(dateFormat)
	sb.WriteString("Created by " + creatorLink + " on " + date)

	// Add a reply view link to each top level reply
	if !post.IsThread() {
		sb.WriteString(", " + md.Link("#"+post.ID.String(), makeReplyURI(post)))
	}

	if post.reposts.Size() > 0 {
		sb.WriteString(", " + strconv.Itoa(post.reposts.Size()) + " repost(s)")
	}

	sb.WriteString("  \n")

	actions := []string{
		md.Link("Flag", makeFlagURI(post)),
	}

	if post.IsThread() {
		actions = append(actions, md.Link("Repost", makeCreateRepostURI(post)))
	}

	isReadonly := post.Readonly || post.Board.Readonly
	if !isReadonly {
		actions = append(
			actions,
			md.Link("Reply", makeCreateReplyURI(post)),
			md.Link("Edit", makeEditPostURI(post)),
			md.Link("Delete", makeDeletePostURI(post)),
		)
	}

	if levels == 0 {
		if post.IsThread() {
			actions = append(actions, md.Link("Show all Replies", makeThreadURI(post)))
		} else {
			actions = append(actions, md.Link("View Thread", makeThreadURI(post)))
		}
	}

	sb.WriteString(strings.Join(actions, " • ") + " \n")
}

func (post *Post) Render(path string, indent string, levels int) string {
	if post == nil {
		return ""
	}

	var sb strings.Builder

	// Thread reposts might not have a title, if so get title from source thread
	title := post.Title
	if post.IsRepost() && title == "" {
		if board, ok := getBoard(post.RepostBoardID); ok {
			if src, ok := board.GetThread(post.ParentID); ok {
				title = src.Title
			}
		}
	}

	if title != "" { // Replies don't have a title
		sb.WriteString(md.H1(title))
	}
	sb.WriteString(indent + "\n")

	post.renderPostContent(&sb, indent, levels)

	if post.replies.Size() == 0 {
		return sb.String()
	}

	// XXX: This triggers for reply views
	if levels == 0 {
		sb.WriteString(indent + "\n")
		return sb.String()
	}

	if path != "" {
		sb.WriteString(post.renderTopLevelReplies(path, indent, levels-1))
	} else {
		sb.WriteString(post.renderSubReplies(indent, levels-1))
	}
	return sb.String()
}

func (post *Post) renderTopLevelReplies(path, indent string, levels int) string {
	p, err := pager.New(path, post.replies.Size(), pager.WithPageSize(pageSizeReplies))
	if err != nil {
		panic(err)
	}

	var (
		b              strings.Builder
		commentsIndent = indent + "> "
	)

	render := func(_ string, v any) bool {
		reply := v.(*Post)
		b.WriteString(indent + "\n" + reply.Render("", commentsIndent, levels-1))
		return false
	}

	b.WriteString("\n" + md.HorizontalRule() + "Sort by: ")
	r := parseRealmPath(path)
	if r.Query.Get("order") == "desc" {
		r.Query.Set("order", "asc")
		b.WriteString(md.Link("newest first", r.String()) + "\n")
		post.replies.ReverseIterateByOffset(p.Offset(), p.PageSize(), render)

	} else {
		r.Query.Set("order", "desc")
		b.WriteString(md.Link("oldest first", r.String()) + "\n")
		post.replies.IterateByOffset(p.Offset(), p.PageSize(), render)
	}

	if p.HasPages() {
		b.WriteString(md.HorizontalRule())
		b.WriteString(pager.Picker(p))
	}

	return b.String()
}

func (post *Post) renderSubReplies(indent string, levels int) string {
	var (
		b              strings.Builder
		commentsIndent = indent + "> "
	)

	post.replies.Iterate("", "", func(_ string, v any) bool {
		reply := v.(*Post)
		b.WriteString(indent + "\n" + reply.Render("", commentsIndent, levels-1))
		return false
	})

	return b.String()
}

func (post *Post) RenderInner() string {
	if post.IsThread() {
		panic("unexpected thread")
	}

	var (
		s         string
		threadID  = post.ThreadID
		thread, _ = post.Board.GetThread(threadID) // TODO: This seems redundant (post == thread)
	)

	// Fully render parent if it's not a repost.
	if !post.IsRepost() {
		var (
			parent   *Post
			parentID = post.ParentID
		)

		if thread.ID == parentID {
			parent = thread
		} else {
			parent, _ = thread.GetReply(parentID)
		}

		s += parent.Render("", "", 0) + "\n"
	}

	s += post.Render("", "> ", 5)
	return s
}
