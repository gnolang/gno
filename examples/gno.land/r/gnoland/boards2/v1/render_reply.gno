package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/jeronimoalbi/mdform"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/leon/svgbtn"
	"gno.land/p/moul/md"
	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/ufmt/v0"
)

func renderReply(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write("Board not found")
		return
	}

	rawID := req.GetVar("thread")
	threadID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid thread ID: " + md.EscapeText(rawID))
		return
	}

	rawID = req.GetVar("reply")
	replyID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid reply ID: " + md.EscapeText(rawID))
		return
	}

	thread, found := getThread(board, boards.ID(threadID))
	if !found {
		res.Write("Thread not found")
		return
	}

	reply, found := getReply(thread, boards.ID(replyID))
	if !found {
		res.Write("Reply not found")
		return
	}

	// Call render even for hidden replies to display children.
	// Original comment content will be hidden under the hood.
	// See: #3480
	res.Write(renderPostInner(reply, req.RawPath))
}

// newClampedPager builds a pager, clamping an out-of-range, zero, negative,
// or malformed ?page= to the last valid page instead of erroring. pager.New
// rejects page==0, page>pageCount, and non-numeric pages with
// ErrInvalidPageNumber, which the callers would otherwise surface as a panic
// (aborting the whole render) or an error page in place of the list. A
// NEGATIVE page is not rejected by pager.New (it only checks ==0 and
// >pageCount), so it must be caught here too — otherwise it renders a broken
// "page -N of M" picker with both arrows disabled. Triggers on a stale deep
// link after deletions, or a hand-edited URL.
func newClampedPager(path string, size, pageSize int) pager.Pager {
	p, err := pager.New(path, size, pager.WithPageSize(pageSize))
	if err == nil && p.Page() >= 1 {
		return p
	}
	// Re-parse without ?page= to get a valid page-1 pager and read the real
	// page count, then jump to the last page when there is more than one.
	r := parseRealmPath(path)
	r.Query.Del("page")
	first, ferr := pager.New(r.String(), size, pager.WithPageSize(pageSize))
	if ferr != nil {
		// A page-less path is valid for every current route, so this is
		// unreachable today; fall back to the original pager rather than
		// asserting the invariant with a panic that would abort the render.
		return p
	}
	if last := first.PageCount(); last > 1 {
		r.Query.Set("page", strconv.Itoa(last))
		if clamped, e := pager.New(r.String(), size, pager.WithPageSize(pageSize)); e == nil {
			return clamped
		}
	}
	return first
}

func renderTopLevelReplies(post *boards.Post, path, indent string, levels int, budget *int) string {
	p := newClampedPager(path, post.Replies.Size(), pageSizeReplies)
	link, desc := sortToggleLink(path)

	var (
		b              strings.Builder
		commentsIndent = indent + "> "
		truncated      bool
	)

	render := func(reply *boards.Post) bool {
		if *budget <= 0 {
			truncated = true
			return true // stop: render budget exhausted (see maxRenderedBodies)
		}
		// Forward the page order so this reply's nested children match it.
		b.WriteString(indent + "\n" + renderPost(reply, "", commentsIndent, levels-1, budget, desc))
		return false
	}

	b.WriteString("\n" + md.HorizontalRule() + "Sort by: " + link + "\n")

	count := p.PageSize()
	if desc {
		count = -count // Reverse iterate
	}

	post.Replies.Iterate(p.Offset(), count, render)

	if truncated {
		b.WriteString(indent + "\n" + commentsIndent + "_Some replies not shown — " +
			md.Link("view all comments", makeThreadFlatURI(post)) + "._\n")
	}

	// Suppress the page picker when the budget truncated this page: later
	// replies in the page were skipped, so the offset-based "next page" would
	// jump past them. The flat link above is the complete, reachable view.
	if !truncated && p.HasPages() {
		b.WriteString(md.HorizontalRule())
		b.WriteString(pager.Picker(p))
	}
	return b.String()
}

func renderSubReplies(post *boards.Post, indent string, levels int, budget *int, desc bool) string {
	var (
		b              strings.Builder
		commentsIndent = indent + "> "
		truncated      bool
	)

	// Cap inline children at pageSizeReplies. A nested reply list can't have
	// its own pager (it would collide with the page's ?page=), so instead of
	// dumping every child here a comment with more links to its own re-rooted
	// view, which paginates them. Keeps every view bounded to <=pageSizeReplies
	// children per post. count's sign follows the page order so the inline
	// children match the view they appear in (desc → the newest ones).
	count := pageSizeReplies
	if desc {
		count = -count
	}
	post.Replies.Iterate(0, count, func(reply *boards.Post) bool {
		if *budget <= 0 {
			truncated = true
			return true // stop: render budget exhausted (see maxRenderedBodies)
		}
		b.WriteString(indent + "\n" + renderPost(reply, "", commentsIndent, levels-1, budget, desc))
		return false
	})

	notice := func(text, uri string) {
		b.WriteString(indent + "\n" + commentsIndent + md.Link(text, uri) + "\n")
	}
	switch {
	case truncated:
		// Budget exhausted: the whole-thread flat view is the reachable backstop.
		notice("More replies — view all comments", makeThreadFlatURI(post))
	case post.Replies.Size() > pageSizeReplies:
		// Breadth cap hit: re-root at this comment to page the rest, carrying
		// the page's sort order so the re-root opens the same way (its first
		// page then matches the newest-first children shown inline here).
		uri := makeReplyURI(post)
		if desc {
			uri += "?order=desc"
		}
		notice("View all "+strconv.Itoa(post.Replies.Size())+" replies", uri)
	}
	return b.String()
}

func renderEditReply(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write("Board not found")
		return
	}

	rawID := req.GetVar("thread")
	threadID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid thread ID: " + md.EscapeText(rawID))
		return
	}

	rawID = req.GetVar("reply")
	replyID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid reply ID: " + md.EscapeText(rawID))
		return
	}

	thread, found := getThread(board, boards.ID(threadID))
	if !found {
		res.Write("Thread not found")
		return
	}

	reply, found := getReply(thread, boards.ID(replyID))
	if !found {
		res.Write("Reply not found")
		return
	}

	form := mdform.New("exec", "EditReply")
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
	form.Input(
		"replyID",
		"placeholder", "Reply ID",
		"value", reply.ID.String(),
		"readonly", "true",
	)
	form.Textarea(
		"body",
		"placeholder", "Comment",
		"value", reply.Body,
		"required", "true",
	)

	res.Write(md.H1(board.Name + ": Edit Comment"))
	res.Write(md.Link("← Back to thread", makeThreadURI(thread)) + "\n\n")
	res.Write(
		md.Paragraph(
			ufmt.Sprintf("Editing a comment from the thread: %s", md.Link(thread.Title, makeThreadURI(thread))),
		),
	)
	res.Write(form.String())
	res.Write("\n\n**Done?** " + svgbtn.ButtonWithRadius(136, 32, 4, "#E2E2E2", "#54595D", "Return to thread", makeThreadURI(thread)) + "\n")
}
