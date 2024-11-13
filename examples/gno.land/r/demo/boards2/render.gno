package boards

import (
	"strconv"

	"gno.land/p/demo/mux"
)

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderBoardsList)
	router.HandleFunc("{board}", renderBoard)
	router.HandleFunc("{board}/{thread}", renderThread)
	router.HandleFunc("{board}/{thread}/{reply}", renderReply)

	router.NotFoundHandler = func(res *mux.ResponseWriter, _ *mux.Request) {
		res.Write("Path not found")
	}

	return router.Render(path)
}

func renderBoardsList(res *mux.ResponseWriter, _ *mux.Request) {
	res.Write("These are all the boards of this realm:\n\n")
	gBoards.Iterate("", "", func(_ string, value interface{}) bool {
		board := value.(*Board)
		res.Write(" * " + newLink(board.url, board.url) + "\n")
		return false
	})
}

func renderBoard(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	v, found := gBoardsByName.Get(name)
	if !found {
		res.Write("Board does not exist: " + name)
	} else {
		board := v.(*Board)
		res.Write(board.Render())
	}
}

func renderThread(res *mux.ResponseWriter, req *mux.Request) {
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

func renderReply(res *mux.ResponseWriter, req *mux.Request) {
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
