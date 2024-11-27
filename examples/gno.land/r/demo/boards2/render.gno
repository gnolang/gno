package boards2

import (
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
)

const (
	defaultPageSize   = 50
	threadRenderDepth = 5
)

func Render(path string) string {
	// TODO: remove this when query strings PR will be merged: https://github.com/gnolang/gno/pull/2594
	router := newWrappedRouter()
	router.HandleFunc("", renderBoardsList)
	router.HandleFunc("{board}", renderBoard)
	router.HandleFunc("{board}/{thread}", renderThread)
	router.HandleFunc("{board}/{thread}/{reply}", renderReply)

	router.NotFoundHandler = func(res *mux.ResponseWriter, _ *mux.Request) {
		res.Write("Path not found")
	}

	return router.Render(path)
}

func renderBoardsList(res *mux.ResponseWriter, r request) {
	res.Write("These are all the boards of this realm:\n\n")

	pg := pager.NewPager(&gBoardsByID, defaultPageSize)
	pageNum, limit, err := pg.ParseQuery(r.requestUri)
	if err != nil {
		panic(err)
	}

	page := pg.GetPageWithSize(pageNum, limit)
	for _, item := range page.Items {
		board := item.Value.(*Board)
		url := board.GetURL()
		res.Write(" * " + newLink(url, url) + "\n")
	}

	res.Write("\n\n---\n\n")
	res.Write(page.Selector())
}

func renderBoard(res *mux.ResponseWriter, req request) {
	name := req.GetVar("board")
	v, found := gBoardsByName.Get(name)
	if !found {
		res.Write("Board does not exist: " + name)
		return
	}

	board := v.(*Board)
	p := req.pagination(defaultPageSize)
	res.Write(board.Render(p))
}

func renderThread(res *mux.ResponseWriter, req request) {
	name := req.GetVar("board")
	v, found := gBoardsByName.Get(name)
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

	board := v.(*Board)
	thread, found := board.GetThread(PostID(tID))
	if !found {
		res.Write("Thread does not exist with ID: " + req.GetVar("thread"))
		return
	}
	res.Write(thread.Render("", threadRenderDepth))
}

func renderReply(res *mux.ResponseWriter, req request) {
	name := req.GetVar("board")
	v, found := gBoardsByName.Get(name)
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

	rawID = req.GetVar("reply")
	rID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid reply ID: " + rawID)
		return
	}

	board := v.(*Board)
	thread, found := board.GetThread(PostID(tID))
	if !found {
		res.Write("Thread does not exist with ID: " + req.GetVar("thread"))
		return
	}

	reply, found := thread.GetReply(PostID(rID))
	if !found {
		res.Write("Reply does not exist with ID: " + req.GetVar("reply"))
	} else {
		res.Write(reply.RenderInner())
	}
}
