package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/jeronimoalbi/forms"
	"gno.land/p/moul/md"
	"gno.land/p/nt/mux"
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

	b.WriteString(md.Bold("≡ "+md.Link(summary, postURI)) + "  \n")
	b.WriteString("Created by " + creatorLink + " on " + date + "  \n")

	status := []string{
		strconv.Itoa(thread.Replies.Size()) + " replies",
		strconv.Itoa(thread.Reposts.Size()) + " reposts",
	}
	b.WriteString(md.Bold(strings.Join(status, " • ")) + "\n")
	return b.String()
}

func renderFlagPost(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write("Board does not exist: " + name)
		return
	}

	// Thread ID must always be available
	threadID := req.GetVar("thread")
	if _, err := strconv.Atoi(threadID); err != nil {
		res.Write("Invalid thread ID: " + threadID)
		return
	}

	// Parse reply ID when post is a reply
	replyID := req.GetVar("reply")
	if replyID != "" {
		_, err := strconv.Atoi(replyID)
		if err != nil {
			res.Write("Invalid reply ID: " + replyID)
			return
		}
	}

	exec := "FlagThread"
	if replyID != "" {
		exec = "FlagReply"
	}

	form := forms.New("exec", exec)
	form.Input(
		"boardID",
		"placeholder", "Board ID",
		"value", board.ID.String(),
		"readonly", "true",
	)
	form.Input(
		"threadID",
		"placeholder", "Thread ID",
		"value", threadID,
		"readonly", "true",
	)

	if replyID != "" {
		form.Input(
			"replyID",
			"placeholder", "Reply ID",
			"value", replyID,
		)
	}

	form.Input(
		"reason",
		"placeholder", "Reason",
	)

	if replyID == "" {
		res.Write(md.H1(board.Name + ": Flag Thread"))
	} else {
		res.Write(md.H1(board.Name + ": Flag Reply"))
	}

	res.Write(form.String())
}
