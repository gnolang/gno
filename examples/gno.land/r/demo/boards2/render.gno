package boards2

import (
	"strconv"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
)

const defaultPageSize = 50

func Render(path string) string {
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

	pager := pager.NewPager(&gBoardsByID, defaultPageSize)
	pageNum, pageSize, err := pager.ParseQuery(r.requestUri)
	if err != nil {
		panic(err)
	}

	page := pager.GetPageWithSize(pageNum, pageSize)
	for _, item := range page.Items {
		board := item.Value.(*Board)
		url := board.GetURL()
		res.Write(" * " + newLink(url, url) + "\n")
	}

	res.Write(page.Selector())
}

func renderBoard(res *mux.ResponseWriter, req request) {
	name := req.GetVar("board")
	v, found := gBoardsByName.Get(name)
	if !found {
		res.Write("Board does not exist: " + name)
	} else {
		board := v.(*Board)
		res.Write(board.Render())
	}
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
	} else {
		res.Write(thread.Render("", 5))
	}
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
