package boards2

import (
	"net/url"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/mux"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
)

const (
	pageSizeBoards  = 6
	pageSizeThreads = 6
	pageSizeReplies = 10
)

const menuManageBoard = "manageBoard"

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderBoardsList)
	router.HandleFunc("help", renderHelp)
	router.HandleFunc("admin-users", renderMembers)
	router.HandleFunc("{board}", renderBoard)
	router.HandleFunc("{board}/members", renderMembers)
	router.HandleFunc("{board}/{thread}", renderThread)
	router.HandleFunc("{board}/{thread}/{reply}", renderReply)

	router.NotFoundHandler = func(res *mux.ResponseWriter, _ *mux.Request) {
		res.Write("Path not found")
	}

	return router.Render(path)
}

func renderHelp(res *mux.ResponseWriter, _ *mux.Request) {
	res.Write("# Boards Help\n")
	if gHelp != "" {
		res.Write(gHelp)
	} else {
		res.Write("### Help content has not been uploaded\n")
		res.Write("Do you want to ")
		res.Write(newLink("upload boards help", txlink.Call("SetHelp")))
		res.Write(" ?")
	}
}

func renderBoardsList(res *mux.ResponseWriter, req *mux.Request) {
	renderNotice(res)

	res.Write("# Boards\n\n")
	renderBoardListMenu(res, req)
	res.Write("\n\n---\n")

	if gBoardsByID.Size() == 0 {
		res.Write("### Currently there are no boards\n")
		res.Write("Be the first to " + newLink("create a new board", txlink.Call("CreateBoard")) + " !")

		return
	}

	p, err := pager.New(req.RawPath, gBoardsByID.Size(), pager.WithPageSize(pageSizeBoards))
	if err != nil {
		panic(err)
	}

	render := func(_ string, v any) bool {
		board := v.(*Board)

		res.Write("## " + newLink(board.GetName(), board.GetURL()) + "  \n")
		res.Write("_Created by " + newUserLink(board.creator) + " ")
		res.Write("on " + board.createdAt.Format(dateFormat))
		res.Write(", #" + board.id.String() + "_  \n")
		res.Write("_" + strconv.Itoa(board.threads.Size()) + " threads")
		if board.IsReadonly() {
			res.Write(", **read-only**")
		}

		res.Write("_\n")
		return false
	}

	res.Write("_Sort by ")
	r := realmpath.Parse(req.RawPath)
	if r.Query.Get("order") == "desc" {
		r.Query.Set("order", "asc")
		res.Write(newLink("newest first", r.String()))
		res.Write("_\n")
		gBoardsByID.ReverseIterateByOffset(p.Offset(), p.PageSize(), render)
	} else {
		r.Query.Set("order", "desc")
		res.Write(newLink("oldest first", r.String()))
		res.Write("_\n")
		gBoardsByID.IterateByOffset(p.Offset(), p.PageSize(), render)
	}

	if p.HasPages() {
		res.Write("\n---\n")
		res.Write(pager.Picker(p))
	}
}

func renderBoardListMenu(res *mux.ResponseWriter, req *mux.Request) {
	path := strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land")

	res.Write(newLink("Create Board", txlink.Call("CreateBoard")))
	res.Write(" • ")
	res.Write(newLink("List Admin Users", path+":admin-users"))
	res.Write(" • ")
	res.Write(newLink("Help", path+":help"))
	res.Write("\n\n")
}

func renderBoard(res *mux.ResponseWriter, req *mux.Request) {
	renderNotice(res)

	name := req.GetVar("board")
	v, found := gBoardsByName.Get(name)
	if !found {
		res.Write("### The board you are looking for does not exist\n")
		res.Write("Do you want to [create a new board](" + txlink.Call("CreateBoard", "name", name) + ") ?")
		return
	}

	board := v.(*Board)
	menu := renderBoardMenu(board, req)

	res.Write(board.Render(req.RawPath, menu))
}

func renderBoardMenu(board *Board, req *mux.Request) string {
	var (
		b               strings.Builder
		boardMembersURL = board.GetPath() + "/members"
	)

	if board.IsReadonly() {
		b.WriteString(newLink("List Members", boardMembersURL))
		b.WriteString(" • ")
		b.WriteString(newLink("Unfreeze Board", board.GetUnfreezeFormURL()))
		b.WriteString("\n")
	} else {
		b.WriteString(newLink("Create Thread", board.GetPostFormURL()))
		b.WriteString(" • ")

		menu := getCurrentMenu(req.RawPath)
		if menu == menuManageBoard {
			b.WriteString("**Manage Board**")
		} else {
			b.WriteString(newLink("Manage Board", menuURL(menuManageBoard)))
		}

		b.WriteString("  \n")

		if menu == menuManageBoard {
			b.WriteString("↳")
			b.WriteString(newLink("Invite Member", board.GetInviteMemberFormURL()))
			b.WriteString(" • ")
			b.WriteString(newLink("List Members", boardMembersURL))
			b.WriteString(" • ")
			b.WriteString(newLink("Freeze Board", board.GetFreezeFormURL()))
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

func renderThread(res *mux.ResponseWriter, req *mux.Request) {
	renderNotice(res)

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
		res.Write(thread.Render(req.RawPath, "", 5))
	}
}

func renderReply(res *mux.ResponseWriter, req *mux.Request) {
	renderNotice(res)

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
	boardID := BoardID(0)
	perms := gPerms
	name := req.GetVar("board")
	if name != "" {
		v, found := gBoardsByName.Get(name)
		if !found {
			res.Write("### Board not found")
			return
		}

		board := v.(*Board)
		boardID = board.GetID()
		perms = board.perms

		res.Write("# " + board.GetName() + " Members\n")
		res.Write("### These are the board members\n\n")
	} else {
		res.Write("# Admin Users\n")
		res.Write("### These are the admin users of the realm\n\n")
	}

	// Create a pager with a small page size to reduce
	// the number of username lookups per page.
	p, err := pager.New(req.RawPath, perms.UsersCount(), pager.WithPageSize(5))
	if err != nil {
		res.Write(err.Error())
		return
	}

	res.Write("| Member | Role | Actions |\n| --- | --- | --- |\n")
	perms.IterateUsers(p.Offset(), p.PageSize(), func(u User) bool {
		res.Write("| ")
		res.Write(newUserLink(u.Address))
		res.Write(" | ")
		res.Write(rolesToString(u.Roles))
		res.Write(" | ")
		res.Write(newLink("remove", txlink.Call(
			"RemoveMember",
			"boardID", boardID.String(),
			"member", u.Address.String(),
		)))
		res.Write(" • ")
		res.Write(newLink("change role", txlink.Call(
			"ChangeMemberRole",
			"boardID", boardID.String(),
			"member", u.Address.String(),
		)))
		res.Write(" |\n")
		return false
	})

	if p.HasPages() {
		res.Write("\n\n" + pager.Picker(p))
	}
}

func renderNotice(res *mux.ResponseWriter) {
	if gNotice == "" {
		return
	}

	res.Write("> ")
	res.Write(gNotice)
	res.Write("\n\n")
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

func menuURL(name string) string {
	// TODO: Menu URL works because no other GET arguments are being used
	return "?menu=" + name
}

func getCurrentMenu(rawURL string) string {
	_, rawQuery, found := strings.Cut(rawURL, "?")
	if !found {
		return ""
	}

	query, _ := url.ParseQuery(rawQuery)
	return query.Get("menu")
}
