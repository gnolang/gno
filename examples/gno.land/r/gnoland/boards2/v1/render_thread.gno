package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/gnoland/boards"
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
	tID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid thread ID: " + rawID)
		return
	}

	thread, found := board.Threads.Get(boards.ID(tID))
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
