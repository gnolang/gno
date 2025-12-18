package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/jeronimoalbi/mdform"
	"gno.land/p/moul/md"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
)

func renderThread(res *mux.ResponseWriter, req *mux.Request) {
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

	thread, found := board.Threads.Get(boards.ID(threadID))
	if !found {
		res.Write("Thread does not exist with ID: " + rawID)
		return
	}

	if thread.Hidden {
		res.Write("Thread with ID: " + rawID + " has been flagged as inappropriate")
		return
	}

	res.Write(md.H1("Boards › " + board.Name))
	res.Write(renderPost(thread, req.RawPath, "", 5))
}

func renderThreadSummary(thread *boards.Post) string {
	var (
		b           strings.Builder
		postURI     = makeThreadURI(thread)
		summary     = summaryOf(thread.Title, 80)
		creatorLink = userLink(thread.Creator)
		date        = thread.CreatedAt.Format(dateFormat)
	)

	b.WriteString(md.Bold(md.Link(summary, postURI)) + "  \n")
	b.WriteString("Created by " + creatorLink + " on " + date + "  \n")

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
		res.Write("Board does not exist: " + name)
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
	)
	form.Textarea(
		"body",
		"placeholder", "Content",
		"rows", "10",
	)

	res.Write(md.H1(board.Name + ": Create Thread"))
	res.Write(
		md.Paragraph(
			ufmt.Sprintf("Thread will be created in %s board", md.Link(board.Name, makeBoardURI(board))),
		),
	)
	res.Write(form.String())
}

func renderEditThread(res *mux.ResponseWriter, req *mux.Request) {
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

	thread, found := board.Threads.Get(boards.ID(threadID))
	if !found {
		res.Write("Thread does not exist with ID: " + rawID)
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
	)
	form.Textarea(
		"bod",
		"placeholder", "Content",
		"rows", "10",
		"value", thread.Body,
	)

	res.Write(md.H1(board.Name + ": Edit Thread"))
	res.Write(
		md.Paragraph("Editing " + md.Link(thread.Title, makeThreadURI(thread))),
	)
	res.Write(form.String())
}

func renderRepostThread(res *mux.ResponseWriter, req *mux.Request) {
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

	thread, found := board.Threads.Get(boards.ID(threadID))
	if !found {
		res.Write("Thread does not exist with ID: " + rawID)
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
	)
	form.Input(
		"title",
		"placeholder", "Title",
	)
	form.Textarea(
		"body",
		"placeholder", "Content",
		"rows", "10",
	)

	res.Write(md.H1(board.Name + ": Repost Thread"))
	res.Write(
		md.Paragraph(
			"Threads can be reposted to other open boards or boards where you are a member " +
				"and are allowed to create new threads.",
		),
	)
	res.Write(
		md.Paragraph(
			ufmt.Sprintf("Reposting %s thread.", md.Link(thread.Title, makeThreadURI(thread))),
		),
	)
	res.Write(form.String())
}
