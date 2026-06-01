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

// maxFlatIndentDepth caps the blockquote nesting in the flat comment view so
// deep reply chains stay readable; comments deeper than this still render,
// just at the capped indent.
const maxFlatIndentDepth = 6

func renderThread(res *mux.ResponseWriter, req *mux.Request) {
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

	thread, found := getThread(board, boards.ID(threadID))
	if !found {
		res.Write("Thread not found")
		return
	}

	if thread.Hidden {
		link := md.Link("inappropriate", makeFlaggingReasonsURI(thread))
		res.Write("⚠ Thread has been flagged as " + link)
		return
	}

	res.Write(md.H1(md.Link("Boards", gRealmPath) + " › " + md.Link(board.Name, makeBoardURI(board))))
	budget := maxRenderedBodies()
	if parseRealmPath(req.RawPath).Query.Get("flat") != "" {
		res.Write(renderThreadFlat(thread, req.RawPath, &budget))
		return
	}
	res.Write(renderPost(thread, req.RawPath, "", 5, &budget, false))
}

// renderThreadFlat renders every comment in the thread as a single flat,
// depth-indented, paginated list backed by ThreadMeta.AllReplies (the index
// that already holds every reply at every depth). Unlike the recursive
// threaded view — which bounds work with the render budget and truncates a
// large subtree — this view is reachable to the very last comment: each
// comment renders one <gno-foreign> block (no recursion), so a fixed page
// size (pageSizeFlat) plus the OP stays well under the budget regardless of
// nesting, and the pager always advances. ?order=desc shows newest first, so
// its page 1 is the latest comments.
func renderThreadFlat(thread *boards.Post, path string, budget *int) string {
	var b strings.Builder

	// The OP for context (its body only; no replies — levels 0).
	b.WriteString(renderPost(thread, "", "", 0, budget, false))

	meta, ok := thread.Meta.(*ThreadMeta)
	if !ok || meta.AllReplies.Size() == 0 {
		return b.String()
	}
	all := meta.AllReplies
	p := newClampedPager(path, all.Size(), pageSizeFlat)

	b.WriteString("\n" + md.HorizontalRule())
	b.WriteString(md.Link("← Threaded view", makeThreadURI(thread)) + " · All " +
		strconv.Itoa(all.Size()) + " comments — sort by: ")

	// sortToggleLink preserves flat=1, so the toggle stays in the flat view.
	link, desc := sortToggleLink(path)
	b.WriteString(link + "\n")

	count := p.PageSize()
	if desc {
		count = -count // reverse iterate: newest first
	}
	all.Iterate(p.Offset(), count, func(reply *boards.Post) bool {
		if *budget <= 0 {
			// Unreachable while pageSizeFlat << maxRenderedBodies; a backstop
			// in case a chain upgrade drops the native cap below one page.
			return true
		}
		indent := flatIndent(thread, reply)
		b.WriteString(indent + "\n" + renderPost(reply, "", indent, 0, budget, false))
		return false
	})

	if p.HasPages() {
		b.WriteString(md.HorizontalRule())
		b.WriteString(pager.Picker(p))
	}
	return b.String()
}

// flatIndent returns the blockquote indent for a reply in the flat view,
// derived from its depth below the thread root (depth 1 = a direct reply to
// the thread), capped at maxFlatIndentDepth.
func flatIndent(thread *boards.Post, reply *boards.Post) string {
	depth := 1
	pid := reply.ParentID
	for pid != thread.ID && pid != 0 && depth < maxFlatIndentDepth {
		parent, ok := getReply(thread, pid)
		if !ok {
			break
		}
		depth++
		pid = parent.ParentID
	}
	return strings.Repeat("> ", depth)
}

func renderThreadSummary(thread *boards.Post) string {
	var (
		b           strings.Builder
		postURI     = makeThreadURI(thread)
		summary     = summaryOf(thread.Title, 80)
		creatorLink = userLink(thread.Creator)
		roleBadge   = getRoleBadge(thread)
		date        = thread.CreatedAt.Format(dateFormat)
	)

	if boards.IsRepost(thread) {
		summary += ` ⟳`
		postURI += ` "This is a thread repost"`
	}

	b.WriteString(md.H6(md.Link(summary, postURI)))
	b.WriteString("Created by " + creatorLink + roleBadge + " on " + date + "  \n")

	status := []string{
		strconv.Itoa(thread.Replies.Size()) + " replies",
		strconv.Itoa(thread.Reposts.Size()) + " reposts",
	}
	b.WriteString(md.Bold(strings.Join(status, " • ")) + "\n")
	return b.String()
}

func renderCreateThread(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write("Board not found")
		return
	}

	form := mdform.New("exec", "CreateThread")
	form.Input(
		"boardID",
		"placeholder", "Board ID",
		"value", board.ID.String(),
		"readonly", "true",
	)
	form.Input(
		"title",
		"placeholder", "Title",
		"required", "true",
	)
	form.Textarea(
		"body",
		"placeholder", "Content",
		"rows", "10",
		"required", "true",
	)

	res.Write(md.H1(board.Name + ": Create Thread"))
	res.Write(md.Link("← Back to board", makeBoardURI(board)) + "\n\n")
	res.Write(
		md.Paragraph(
			ufmt.Sprintf("Thread will be created in the board: %s", md.Link(board.Name, makeBoardURI(board))),
		),
	)
	res.Write(form.String())
	res.Write("\n\n**Done?** " + svgbtn.ButtonWithRadius(136, 32, 4, "#E2E2E2", "#54595D", "Return to board", makeBoardURI(board)) + "\n")
}

func renderEditThread(res *mux.ResponseWriter, req *mux.Request) {
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

	thread, found := getThread(board, boards.ID(threadID))
	if !found {
		res.Write("Thread not found")
		return
	}

	form := mdform.New("exec", "EditThread")
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
		"title",
		"placeholder", "Title",
		"value", thread.Title,
		"required", "true",
	)
	form.Textarea(
		"body",
		"placeholder", "Content",
		"rows", "10",
		"value", thread.Body,
		"required", "true",
	)

	res.Write(md.H1(board.Name + ": Edit Thread"))
	res.Write(md.Link("← Back to thread", makeThreadURI(thread)) + "\n\n")
	res.Write(
		md.Paragraph("Editing " + md.Link(thread.Title, makeThreadURI(thread))),
	)
	res.Write(form.String())
	res.Write("\n\n**Done?** " + svgbtn.ButtonWithRadius(136, 32, 4, "#E2E2E2", "#54595D", "Return to thread", makeThreadURI(thread)) + "\n")
}

func renderRepostThread(res *mux.ResponseWriter, req *mux.Request) {
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

	thread, found := getThread(board, boards.ID(threadID))
	if !found {
		res.Write("Thread not found")
		return
	}

	form := mdform.New("exec", "CreateRepost")
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
		"destinationBoardID",
		"type", mdform.InputTypeNumber,
		"placeholder", "Board ID where to repost",
		"required", "true",
	)
	form.Input(
		"title",
		"value", thread.Title,
		"placeholder", "Title",
		"required", "true",
	)
	form.Textarea(
		"body",
		"placeholder", "Content",
		"rows", "10",
	)

	res.Write(md.H1(board.Name + ": Repost Thread"))
	res.Write(md.Link("← Back to thread", makeThreadURI(thread)) + "\n\n")
	res.Write(
		md.Paragraph(
			"Threads can be reposted to other open boards or boards where you are a member " +
				"and are allowed to create new threads.",
		),
	)
	res.Write(
		md.Paragraph(
			ufmt.Sprintf("Reposting the thread: %s.", md.Link(thread.Title, makeThreadURI(thread))),
		),
	)
	res.Write(form.String())
	res.Write("\n\n**Done?** " + svgbtn.ButtonWithRadius(136, 32, 4, "#E2E2E2", "#54595D", "Return to thread", makeThreadURI(thread)) + "\n")
}
