package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/jeronimoalbi/mdform"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/md"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
)

func renderReply(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write("Board does not exist: " + name)
		return
	}

	rawID := req.GetVar("thread")
	threadID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid thread ID: " + rawID)
		return
	}

	rawID = req.GetVar("reply")
	replyID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid reply ID: " + rawID)
		return
	}

	thread, found := board.Threads.Get(boards.ID(threadID))
	if !found {
		res.Write("Thread does not exist with ID: " + req.GetVar("thread"))
		return
	}

	reply, found := thread.Replies.Get(boards.ID(replyID))
	if !found {
		res.Write("Reply does not exist with ID: " + rawID)
		return
	}

	// Call render even for hidden replies to display children.
	// Original comment content will be hidden under the hood.
	// See: #3480
	res.Write(renderPostInner(reply))
}

func renderTopLevelReplies(post *boards.Post, path, indent string, levels int) string {
	p, err := pager.New(path, post.Replies.Size(), pager.WithPageSize(pageSizeReplies))
	if err != nil {
		panic(err)
	}

	var (
		b              strings.Builder
		commentsIndent = indent + "> "
	)

	render := func(reply *boards.Post) bool {
		b.WriteString(indent + "\n" + renderPost(reply, "", commentsIndent, levels-1))
		return false
	}

	b.WriteString("\n" + md.HorizontalRule() + "Sort by: ")

	r := parseRealmPath(path)
	sortOrder := r.Query.Get("order")
	if sortOrder == "desc" {
		r.Query.Set("order", "asc")
		b.WriteString(md.Link("newest first", r.String()) + "\n")

	} else {
		r.Query.Set("order", "desc")
		b.WriteString(md.Link("oldest first", r.String()) + "\n")
	}

	count := p.PageSize()
	if sortOrder == "desc" {
		count = -count // Reverse iterate
	}

	post.Replies.Iterate(p.Offset(), count, render)

	if p.HasPages() {
		b.WriteString(md.HorizontalRule())
		b.WriteString(pager.Picker(p))
	}
	return b.String()
}

func renderSubReplies(post *boards.Post, indent string, levels int) string {
	var (
		b              strings.Builder
		commentsIndent = indent + "> "
	)

	post.Replies.Iterate(0, post.Replies.Size(), func(reply *boards.Post) bool {
		b.WriteString(indent + "\n" + renderPost(reply, "", commentsIndent, levels-1))
		return false
	})
	return b.String()
}

func renderEditReply(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write("Board does not exist: " + name)
		return
	}

	rawID := req.GetVar("thread")
	threadID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid thread ID: " + rawID)
		return
	}

	rawID = req.GetVar("reply")
	replyID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid reply ID: " + rawID)
		return
	}

	thread, found := board.Threads.Get(boards.ID(threadID))
	if !found {
		res.Write("Thread does not exist with ID: " + req.GetVar("thread"))
		return
	}

	reply, found := thread.Replies.Get(boards.ID(replyID))
	if !found {
		res.Write("Reply does not exist with ID: " + rawID)
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
		"bod",
		"placeholder", "Comment",
		"value", reply.Body,
	)

	res.Write(md.H1(board.Name + ": Edit Comment"))
	res.Write(
		md.Paragraph(
			ufmt.Sprintf("Editing comment from %s thread", md.Link(thread.Title, makeThreadURI(thread))),
		),
	)
	res.Write(form.String())
}
