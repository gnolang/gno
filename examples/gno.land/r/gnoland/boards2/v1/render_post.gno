package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/jeronimoalbi/mdform"
	"gno.land/p/leon/svgbtn"
	"gno.land/p/moul/md"
	"gno.land/p/nt/mdalert"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
)

func renderPost(post *boards.Post, path, indent string, levels int) string {
	var b strings.Builder

	// Thread reposts might not have a title, if so get title from source thread
	title := post.Title
	if boards.IsRepost(post) && title == "" {
		if board, ok := gBoards.Get(post.OriginalBoardID); ok {
			if src, ok := board.Threads.Get(post.ParentID); ok {
				title = src.Title
			}
		}
	}

	if title != "" { // Replies don't have a title
		b.WriteString(md.H2(title))
	}

	b.WriteString(indent + "\n")
	b.WriteString(renderPostContent(post, indent, levels))

	if post.Replies.Size() == 0 {
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

func renderPostContent(post *boards.Post, indent string, levels int) string {
	var b strings.Builder

	// Author and date header
	creatorLink := userLink(post.Creator)
	roleBadge := getRoleBadge(post)
	date := post.CreatedAt.Format(dateFormat)
	b.WriteString(indent)
	b.WriteString(md.Bold(creatorLink) + roleBadge + " · " + date)
	if !boards.IsThread(post) {
		b.WriteString(" " + md.Link("#"+post.ID.String(), makeReplyURI(post)))
	}
	b.WriteString("  \n")

	// Flagged comment should be hidden, but replies still visible (see: #3480)
	// Flagged threads will be hidden by render function caller.
	if post.Hidden {
		b.WriteString(indentBody(indent, md.Italic("⚠ Reply is hidden as it has been flagged as inappropriate")))
		b.WriteString("\n")
		return b.String()
	}

	srcContent, srcPost := renderSourcePost(post, indent, post.Body == "")
	if boards.IsRepost(post) && srcPost != nil {
		msg := "Original thread is: " + md.Link(srcPost.Title, makeThreadURI(srcPost))
		b.WriteString(mdalert.New(mdalert.TypeInfo, "Thread Repost", msg, true).String())
		b.WriteString("\n")
	}

	// Render repost body before original thread's body
	if post.Body != "" {
		b.WriteString(indentBody(indent, post.Body) + "\n")
		if srcContent != "" {
			// Add extra line to separate repost content from original thread content
			b.WriteString("\n")
		}
	}

	b.WriteString(srcContent)

	// Add a newline to separate source deleted message from repost body content
	if boards.IsRepost(post) && srcPost == nil && len(post.Body) > 0 {
		b.WriteString("\n\n")
	}

	// Split thread content and actions
	if boards.IsThread(post) && !boards.IsRepost(post) {
		b.WriteString("\n")
	}

	// Action buttons
	b.WriteString(indent)
	if !boards.IsThread(post) { // is comment
		b.WriteString("  \n")
		b.WriteString(indent)
	}

	actions := []string{
		md.Link("Flag", makeFlagURI(post)),
	}

	if boards.IsThread(post) {
		repostAction := md.Link("Repost", makeCreateRepostURI(post))
		if post.Reposts.Size() > 0 {
			repostAction += " [" + strconv.Itoa(post.Reposts.Size()) + "]"
		}
		actions = append(actions, repostAction)
	}

	isReadonly := post.Readonly || post.Board.Readonly
	if !isReadonly {
		replyLabel := "Reply"
		if boards.IsThread(post) {
			replyLabel = "Comment"
		}
		replyAction := md.Link(replyLabel, makeCreateReplyURI(post))
		// Add reply count if any
		if post.Replies.Size() > 0 {
			replyAction += " [" + strconv.Itoa(post.Replies.Size()) + "]"
		}

		actions = append(
			actions,
			replyAction,
			md.Link("Edit", makeEditPostURI(post)),
			md.Link("Delete", makeDeletePostURI(post)),
		)
	}

	if levels == 0 {
		if boards.IsThread(post) {
			actions = append(actions, md.Link("Show all Replies", makeThreadURI(post)))
		} else {
			actions = append(actions, md.Link("View Thread", makeThreadURI(post)))
		}
	}

	b.WriteString("↳ " + strings.Join(actions, " • ") + "\n")
	return b.String()
}

func renderPostInner(post *boards.Post) string {
	if boards.IsThread(post) {
		return ""
	}

	var (
		s         string
		threadID  = post.ThreadID
		thread, _ = post.Board.Threads.Get(threadID)
	)

	// Fully render parent if it's not a repost.
	if !boards.IsRepost(post) {
		parentID := post.ParentID
		parent := thread

		if thread.ID != parentID {
			parent, _ = thread.Replies.Get(parentID)
		}

		s += renderPost(parent, "", "", 0) + "\n"
	}

	s += renderPost(post, "", "> ", 5)
	return s
}

func renderSourcePost(post *boards.Post, indent string, inline bool) (string, *boards.Post) {
	if !boards.IsRepost(post) {
		return "", nil
	}

	if !inline {
		indent += "> "
	}

	// TODO: figure out a way to decouple posts from a global storage.
	board, ok := gBoards.Get(post.OriginalBoardID)
	if !ok {
		// TODO: Boards can't be deleted so this might be redundant
		return indentBody(indent, md.Italic("⚠ Source board has been deleted")), nil
	}

	srcPost, ok := board.Threads.Get(post.ParentID)
	if !ok {
		return indentBody(indent, md.Italic("⚠ Source post has been deleted")), nil
	}

	if srcPost.Hidden {
		return indentBody(indent, md.Italic("⚠ Source post has been flagged as inappropriate")), nil
	}

	return indentBody(indent, srcPost.Body) + "\n\n", srcPost
}

func renderFlagPost(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write("Board does not exist: " + name)
		return
	}

	// Thread ID must always be available
	rawID := req.GetVar("thread")
	threadID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid thread ID: " + rawID)
		return
	}

	thread, found := board.Threads.Get(boards.ID(threadID))
	if !found {
		res.Write("Thread does not exist with ID: " + rawID)
		return
	}

	// Parse reply ID when post is a reply
	var reply *boards.Post
	rawID = req.GetVar("reply")
	isReply := rawID != ""
	if isReply {
		replyID, err := strconv.Atoi(rawID)
		if err != nil {
			res.Write("Invalid reply ID: " + rawID)
			return
		}

		reply, _ = thread.Replies.Get(boards.ID(replyID))
		if reply == nil {
			res.Write("Reply does not exist with ID: " + rawID)
			return
		}
	}

	exec := "FlagThread"
	if isReply {
		exec = "FlagReply"
	}

	form := mdform.New("exec", exec)
	form.Input(
		"boardID",
		"placeholder", "Board ID",
		"value", board.ID.String(),
		"readonly", "true",
	)
	form.Input(
		"threadID",
		"placeholder", "Thread ID",
		"value", thread.ID.String(),
		"readonly", "true",
	)

	if isReply {
		form.Input(
			"replyID",
			"placeholder", "Reply ID",
			"value", reply.ID.String(),
			"readonly", "true",
		)
	}

	form.Input(
		"reason",
		"placeholder", "Flagging Reason",
	)

	// Breadcrumb navigation
	backLink := md.Link("← Back to thread", makeThreadURI(thread))

	if isReply {
		res.Write(md.H1(board.Name + ": Flag Comment"))
	} else {
		res.Write(md.H1(board.Name + ": Flag Thread"))
	}
	res.Write(backLink + "\n\n")

	res.Write(
		md.Paragraph(
			"Thread or comment moderation is done though flagging, which is usually done "+
				"by board members with the moderator role, though other roles could also potentially flag.",
		) +
			md.Paragraph(
				"Flagging relies on a configurable threshold, which by default is of one flag, that when "+
					"reached leads to the flagged thread or comment to be hidden.",
			) +
			md.Paragraph(
				"Flagging thresholds can be different within each board.",
			),
	)

	if isReply {
		res.Write(
			md.Paragraph(
				ufmt.Sprintf(
					"⚠ Your are flagging a %s from %s ⚠",
					md.Link("comment", makeReplyURI(reply)),
					userLink(reply.Creator),
				),
			),
		)
	} else {
		res.Write(
			md.Paragraph(
				ufmt.Sprintf(
					"⚠ Your are flagging %s thread ⚠",
					md.Link(thread.Title, makeThreadURI(thread)),
				),
			),
		)
	}

	res.Write(form.String())
	res.Write("\n\n**Done?** " + svgbtn.ButtonWithRadius(136, 32, 4, "#E2E2E2", "#54595D", "Return to thread", makeThreadURI(thread)) + "\n")
}

func renderReplyPost(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write("Board does not exist: " + name)
		return
	}

	// Thread ID must always be available
	rawID := req.GetVar("thread")
	threadID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid thread ID: " + rawID)
		return
	}

	thread, found := board.Threads.Get(boards.ID(threadID))
	if !found {
		res.Write("Thread does not exist with ID: " + rawID)
		return
	}

	// Parse reply ID when post is a reply
	var reply *boards.Post
	rawID = req.GetVar("reply")
	isReply := rawID != ""
	if isReply {
		replyID, err := strconv.Atoi(rawID)
		if err != nil {
			res.Write("Invalid reply ID: " + rawID)
			return
		}

		reply, _ = thread.Replies.Get(boards.ID(replyID))
		if reply == nil {
			res.Write("Reply does not exist with ID: " + rawID)
			return
		}
	}

	form := mdform.New("exec", "CreateReply")
	form.Input(
		"boardID",
		"placeholder", "Board ID",
		"value", board.ID.String(),
		"readonly", "true",
	)
	form.Input(
		"threadID",
		"placeholder", "Thread ID",
		"value", thread.ID.String(),
		"readonly", "true",
	)

	if isReply {
		form.Input(
			"replyID",
			"placeholder", "Reply ID",
			"value", reply.ID.String(),
			"readonly", "true",
		)
	} else {
		form.Input(
			"replyID",
			"placeholder", "Reply ID",
			"value", "0",
			"readonly", "true",
		)
	}

	form.Textarea(
		"body",
		"placeholder", "Comment",
		"required", "true",
	)

	// Breadcrumb navigation
	backLink := md.Link("← Back to thread", makeThreadURI(thread))

	if isReply {
		res.Write(md.H1(board.Name + ": Reply"))
		res.Write(backLink + "\n\n")
		res.Write(
			md.Paragraph(ufmt.Sprintf("Replying to a comment posted by %s:", userLink(reply.Creator))) +
				md.Blockquote(reply.Body),
		)
	} else {
		res.Write(md.H1(board.Name + ": Comment"))
		res.Write(backLink + "\n\n")
		res.Write(
			md.Paragraph(
				ufmt.Sprintf("Commenting on %s thread", md.Link(thread.Title, makeThreadURI(thread))),
			),
		)
	}

	res.Write(form.String())
	res.Write("\n\n**Done?** " + svgbtn.ButtonWithRadius(136, 32, 4, "#E2E2E2", "#54595D", "Return to thread", makeThreadURI(thread)) + "\n")
}
