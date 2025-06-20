package boards2

import (
	"net/url"
	"strconv"
	"strings"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/md"
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
		res.Write(md.Link("upload boards help", gRealmLink.Call("SetHelp", "content", "")))
		res.Write(" ?")
	}
}

func renderBoardsList(res *mux.ResponseWriter, req *mux.Request) {
	renderNotice(res)

	res.Write("# Boards\n\n")
	renderBoardListMenu(res, req)
	res.Write("\n\n---\n")

	boards := gPublicBoardsByID
	if boards.Size() == 0 {
		link := gRealmLink.Call("CreateBoard", "name", "", "private", "false")
		res.Write("### Currently there are no boards\n")
		res.Write("Be the first to ")
		res.Write(md.Link("create a new board", link))
		res.Write(" !")

		return
	}

	p, err := pager.New(req.RawPath, boards.Size(), pager.WithPageSize(pageSizeBoards))
	if err != nil {
		panic(err)
	}

	render := func(_ string, v any) bool {
		board := v.(*Board)

		res.Write("**" + md.Link(board.Name, makeBoardURI(board)) + "**  \n")
		res.Write("Created by " + newUserLink(board.Creator) + " ")
		res.Write("on " + board.CreatedAt().Format(dateFormat))
		res.Write(", #" + board.ID.String() + "  \n")
		res.Write("**" + strconv.Itoa(board.ThreadsCount()) + " threads")
		if board.Readonly {
			res.Write(", read-only")
		}

		res.Write("**\n\n")
		return false
	}

	res.Write("Sort by: ")
	r := parseRealmPath(req.RawPath)
	if r.Query.Get("order") == "desc" {
		r.Query.Set("order", "asc")
		res.Write(md.Link("newest first", r.String()))
		res.Write("\n\n")
		boards.ReverseIterateByOffset(p.Offset(), p.PageSize(), render)
	} else {
		r.Query.Set("order", "desc")
		res.Write(md.Link("oldest first", r.String()))
		res.Write("\n\n")
		boards.IterateByOffset(p.Offset(), p.PageSize(), render)
	}

	if p.HasPages() {
		res.Write("\n---\n")
		res.Write(pager.Picker(p))
	}
}

func renderBoardListMenu(res *mux.ResponseWriter, req *mux.Request) {
	path := strings.TrimPrefix(string(gRealmLink), "gno.land")

	res.Write(md.Link("Create Board", gRealmLink.Call("CreateBoard", "name", "", "private", "false")))
	res.Write(" • ")
	res.Write(md.Link("List Admin Users", path+":admin-users"))
	res.Write(" • ")
	res.Write(md.Link("Help", path+":help"))
	res.Write("\n\n")
}

func renderBoard(res *mux.ResponseWriter, req *mux.Request) {
	renderNotice(res)

	name := req.GetVar("board")
	v, found := gBoardsByName.Get(name)
	if !found {
		link := md.Link("create a new board", gRealmLink.Call("CreateBoard", "name", name, "private", "false"))
		res.Write("### The board you are looking for does not exist\n")
		res.Write("Do you want to ")
		res.Write(link)
		res.Write(" ?")
		return
	}

	board := v.(*Board)
	menu := renderBoardMenu(board, req)

	res.Write(board.Render(req.RawPath, menu))
}

func renderBoardMenu(board *Board, req *mux.Request) string {
	boardMembersURL := ufmt.Sprintf(
		"%s:%s/members",
		strings.TrimPrefix(string(gRealmLink), "gno.land"),
		board.Name,
	)

	var b strings.Builder
	if board.Readonly {
		b.WriteString(md.Link("List Members", boardMembersURL))
		b.WriteString(" • ")
		b.WriteString(md.Link("Unfreeze Board", makeUnfreezeBoardURI(board)))
		b.WriteString("\n")
	} else {
		b.WriteString(md.Link("Create Thread", makeCreateThreadURI(board)))
		b.WriteString(" • ")

		menu := getCurrentMenu(req.RawPath)
		if menu == menuManageBoard {
			b.WriteString("**Manage Board**")
		} else {
			b.WriteString(md.Link("Manage Board", menuURL(menuManageBoard)))
		}

		b.WriteString("  \n")

		if menu == menuManageBoard {
			b.WriteString("↳")
			b.WriteString(md.Link("Invite Member", makeInviteMemberURI(board)))
			b.WriteString(" • ")
			b.WriteString(md.Link("List Members", boardMembersURL))
			b.WriteString(" • ")
			b.WriteString(md.Link("Freeze Board", makeFreezeBoardURI(board)))
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
	} else if thread.Hidden {
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
		boardID = board.ID
		perms = board.perms

		res.Write("# " + board.Name + " Members\n")
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
		res.Write(md.Link("remove", gRealmLink.Call(
			"RemoveMember",
			"boardID", boardID.String(),
			"member", u.Address.String(),
		)))
		res.Write(" • ")
		res.Write(md.Link("change role", gRealmLink.Call(
			"ChangeMemberRole",
			"boardID", boardID.String(),
			"member", u.Address.String(),
			"role", "",
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
