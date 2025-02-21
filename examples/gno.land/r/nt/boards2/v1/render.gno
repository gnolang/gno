package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/demo/mux"
)

const (
	boardsPageSize  = 20
	threadsPageSize = 30
	repliesPageSize = 30
)

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderBoardsList)
	router.HandleFunc("members", renderMembers)
	router.HandleFunc("{board}", renderBoard)
	router.HandleFunc("{board}/members", renderMembers)
	router.HandleFunc("{board}/{thread}", renderThread)
	router.HandleFunc("{board}/{thread}/{reply}", renderReply)

	router.NotFoundHandler = func(res *mux.ResponseWriter, _ *mux.Request) {
		res.Write("Path not found")
	}

	return router.Render(path)
}

func renderBoardsList(res *mux.ResponseWriter, req *mux.Request) {
	res.Write("These are all the boards of this realm:\n\n")
	p := mustGetPagination(req.RawPath, boardsPageSize)
	page := p.Iterate(&gBoardsByID, func(_ string, value interface{}) bool {
		board := value.(*Board)
		path := board.GetPath()
		url := board.GetURL()
		res.Write(" * " + newLink(path, url) + "\n")
		return false
	})

	if page != nil {
		res.Write("\n---\n")
		res.Write(page.Picker())
	}
}

func renderBoard(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	v, found := gBoardsByName.Get(name)
	if !found {
		res.Write("Board does not exist: " + name)
		return
	}

	board := v.(*Board)
	p := mustGetPagination(req.RawPath, threadsPageSize)
	res.Write(board.Render(p))
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
		res.Write("Thread does not exist with ID: " + rawID)
	} else if thread.IsHidden() {
		res.Write("Thread with ID: " + rawID + " has been flagged as inappropriate")
	} else {
		p := mustGetPagination(req.RawPath, repliesPageSize)
		res.Write(thread.Render(p, "", 5))
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
		res.Write("Reply does not exist with ID: " + rawID)
		return
	}

	// Call render even for hidden replies to display children.
	// Original comment content will be hidden under the hood.
	// See: #3480
	res.Write(reply.RenderInner())
}

func renderMembers(res *mux.ResponseWriter, req *mux.Request) {
	perms := gPerm
	name := req.GetVar("board")
	if name != "" {
		v, found := gBoardsByName.Get(name)
		if !found {
			res.Write("Board does not exist: " + name)
			return
		}

		board := v.(*Board)
		perms = board.perms

		res.Write("# Board Members: " + board.GetName() + "\n\n")
	} else {
		res.Write("# Boards Members\n\n")
	}

	// TODO: Add pagination support
	start := 0
	count := 50
	perms.IterateUsers(start, count, func(u User) bool {
		res.Write("- " + u.Address.String() + " " + rolesToString(u.Roles) + "\n")
		return false
	})
}

func rolesToString(roles []Role) string {
	if len(roles) == 0 {
		return ""
	}

	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = string(r)
	}
	return strings.Join(names, ", ")
}
