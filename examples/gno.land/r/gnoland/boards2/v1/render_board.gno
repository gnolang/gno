package boards2

import (
	"gno.land/p/gnoland/boards"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/md"
	"gno.land/p/nt/mux"
)

func renderBoard(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		link := md.Link("create a new board", gRealmLink.Call("CreateBoard", "name", name, "listed", "true"))
		res.Write(md.H3("The board you are looking for does not exist"))
		res.Write("Do you want to " + link + " ?")
		return
	}

	creatorLink := userLink(board.Creator)
	date := board.CreatedAt.Format(dateFormat)

	res.Write(md.H1(board.Name))
	res.Write("Board created by " + creatorLink + " on " + date + ", #" + board.ID.String())
	if board.Readonly {
		res.Write("  \n_" + md.Bold("Starting new threads and commenting is disabled") + "_")
	}

	res.Write("\n" + renderBoardMenu(board, req))
	res.Write(md.HorizontalRule())

	if board.Threads.Size() == 0 {
		res.Write(md.H3("This board doesn't have any threads"))
		if !board.Readonly {
			startConversationLink := md.Link("start a new conversation", makeCreateThreadURI(board))
			res.Write("Do you want to " + startConversationLink + " in this board ?")
		}
		return
	}

	p, err := pager.New(req.RawPath, board.Threads.Size(), pager.WithPageSize(pageSizeDefault))
	if err != nil {
		panic(err)
	}

	render := func(thread *boards.Post) bool {
		if !thread.Hidden {
			res.Write(renderThreadSummary(thread) + "\n")
		}
		return false
	}

	res.Write("Sort by: ")

	r := parseRealmPath(req.RawPath)
	sortOrder := r.Query.Get("order")
	if sortOrder == "desc" {
		r.Query.Set("order", "asc")
		res.Write(md.Link("newest first", r.String()) + "\n\n")
	} else {
		r.Query.Set("order", "desc")
		res.Write(md.Link("oldest first", r.String()) + "\n\n")
	}

	count := p.PageSize()
	if sortOrder == "desc" {
		count = -count // Reverse iterate
	}

	board.Threads.Iterate(p.Offset(), count, render)

	if p.HasPages() {
		res.Write(md.HorizontalRule())
		res.Write(pager.Picker(p))
	}
}
