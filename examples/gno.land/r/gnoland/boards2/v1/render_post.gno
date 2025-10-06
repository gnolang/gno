package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/moul/md"
)

func renderPost(post *Post, path, indent string, levels int) string {
	var b strings.Builder

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
		b.WriteString(md.H1(title))
	}

	b.WriteString(indent + "\n")
	b.WriteString(renderPostContent(post, indent, levels))

	if post.replies.Size() == 0 {
		return b.String()
	}

	// XXX: This triggers for reply views
	if levels == 0 {
		b.WriteString(indent + "\n")
		return b.String()
	}

	if path != "" {
		b.WriteString(renderTopLevelReplies(post, path, indent, levels-1))
	} else {
		b.WriteString(renderSubReplies(post, indent, levels-1))
	}
	return b.String()
}

func renderPostContent(post *Post, indent string, levels int) string {
	var b strings.Builder

	if post.Hidden {
		// Flagged comment should be hidden, but replies still visible (see: #3480)
		// Flagged threads will be hidden by render function caller.
		return indentBody(indent, md.Italic("⚠ Reply is hidden as it has been flagged as inappropriate")) + "\n"
	}

	srcContent, srcPost := renderSourcePost(post, indent)
	if post.IsRepost() && srcPost != nil {
		originLink := md.Link("another thread", makeThreadURI(srcPost))
		b.WriteString("  \nThis thread is a repost of " + originLink + ": \n")
	}

	b.WriteString(srcContent)

	if post.IsRepost() && srcPost == nil && len(post.Body) > 0 {
		// Add a newline to separate source deleted message from repost body content
		b.WriteString("\n")
	}

	b.WriteString(indentBody(indent, post.Body))
	b.WriteString("\n")

	if post.IsThread() {
		// Split content and controls for threads.
		b.WriteString("\n")
	}

	// Buttons & counters
	b.WriteString(indent)
	if !post.IsThread() {
		b.WriteString("  \n")
		b.WriteString(indent)
	}

	creatorLink := userLink(post.Creator)
	date := post.CreatedAt().Format(dateFormat)
	b.WriteString("Created by " + creatorLink + " on " + date)

	// Add a reply view link to each top level reply
	if !post.IsThread() {
		b.WriteString(", " + md.Link("#"+post.ID.String(), makeReplyURI(post)))
	}

	if post.reposts.Size() > 0 {
		b.WriteString(", " + strconv.Itoa(post.reposts.Size()) + " repost(s)")
	}

	b.WriteString("  \n")

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

	b.WriteString(strings.Join(actions, " • ") + " \n")
	return b.String()
}

func renderPostInner(post *Post) string {
	if post.IsThread() {
		return ""
	}

	var (
		s         string
		threadID  = post.ThreadID
		thread, _ = post.Board.GetThread(threadID)
	)

	// Fully render parent if it's not a repost.
	if !post.IsRepost() {
		parentID := post.ParentID
		parent := thread

		if thread.ID != parentID {
			parent, _ = thread.GetReply(parentID)
		}

		s += renderPost(parent, "", "", 0) + "\n"
	}

	s += renderPost(post, "", "> ", 5)
	return s
}

func renderSourcePost(post *Post, indent string) (string, *Post) {
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
