package boards2

import (
	"net/url"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/mux"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/txlink"
)

const (
	boardsPageSize  = 20
	threadsPageSize = 30
	repliesPageSize = 30
)

const (
	menuAdmin      = "admin"
	menuMembership = "membership"
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
	res.Write("# Boards\n\n")
	renderBoardListMenu(res, req)

	res.Write("### These are the boards of this realm\n\n")
	p := mustGetPagination(req.RawPath, boardsPageSize)
	page := p.Iterate(&gBoardsByID, func(_ string, value interface{}) bool {
		board := value.(*Board)

		res.Write(" * " + newLink(board.GetName(), board.GetURL()) + "  \n")
		res.Write("_" + strconv.Itoa(board.threads.Size()) + " threads")
		if board.IsReadOnly() {
			res.Write(", board is read-only")
		}

		res.Write("_\n")
		return false
	})

	if page != nil {
		res.Write("\n---\n")
		res.Write(page.Picker())
	}
}

func renderBoardListMenu(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(newLink("Create Board", txlink.Call("CreateBoard")) + " • ")

	menu := getCurrentSubmenu(req.RawPath)
	if menu == menuMembership {
		res.Write("**Membership**")
	} else {
		res.Write(newLink("Membership", submenuURL(menuMembership)))
	}

	res.Write("  \n")

	if menu == menuMembership {
		path := strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land")

		res.Write("↳")
		res.Write(newLink("Invite", txlink.Call("InviteMember", "boardID", "0")) + " • ")
		res.Write(newLink("Members", path+":members") + " • ")
		res.Write(newLink("Remove Member", txlink.Call("RemoveMember", "boardID", "0")) + " • ")
		res.Write(newLink("Change Member Role", txlink.Call("ChangeMemberRole", "boardID", "0")) + "\n\n")
	}
}

func renderBoard(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	v, found := gBoardsByName.Get(name)
	if !found {
		res.Write("### The board you are looking for does not exist\n")
		res.Write("Do you want to [create a new board](" + txlink.Call("CreateBoard") + ") ?")
		return
	}

	board := v.(*Board)
	renderBoardMenu(board, res, req)

	p := mustGetPagination(req.RawPath, threadsPageSize)
	res.Write(board.Render(p))
}

func renderBoardMenu(board *Board, res *mux.ResponseWriter, req *mux.Request) {
	res.Write(newLink("Start Thread", board.GetPostFormURL()) + " • ")

	menu := getCurrentSubmenu(req.RawPath)
	if menu == menuMembership {
		res.Write("**Membership** • ")
	} else {
		res.Write(newLink("Membership", submenuURL(menuMembership)) + " • ")
	}

	if menu == menuAdmin {
		res.Write("**Admin**")
	} else {
		res.Write(newLink("Admin", submenuURL(menuAdmin)))
	}

	res.Write("  \n")

	if menu != "" {
		res.Write("↳")
	}

	switch menu {
	case menuAdmin:
		res.Write(newLink("Rename Board", board.GetRenameFormURL()) + " • ")
		res.Write(newLink("Freeze", board.GetFreezeFormURL()) + " • ")
		res.Write(newLink("Change Flagging Threshold", board.GetFlaggingThresholdFormURL()) + "\n\n")
	case menuMembership:
		res.Write(newLink("Invite", board.GetInviteMemberFormURL()) + " • ")
		res.Write(newLink("Members", board.GetPath()+"/members") + " • ")
		res.Write(newLink("Remove Member", board.GetRemoveMemberFormURL()) + " • ")
		res.Write(newLink("Change Member Role", board.GetChangeMemberRoleFormURL()) + "\n\n")
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
	perms := gPerms
	name := req.GetVar("board")
	if name != "" {
		v, found := gBoardsByName.Get(name)
		if !found {
			res.Write("### Board not found")
			return
		}

		board := v.(*Board)
		perms = board.perms

		res.Write("# Board Members: " + board.GetName() + "\n\n")
	} else {
		res.Write("# Boards Members\n\n")
	}

	res.Write("### These are the board members\n\n")

	p, err := pager.New(req.RawPath, perms.UsersCount())
	if err != nil {
		res.Write(err.Error())
		return
	}

	perms.IterateUsers(p.Offset(), p.PageSize(), func(u User) bool {
		res.Write("- " + u.Address.String() + " " + rolesToString(u.Roles) + "\n")
		return false
	})

	if p.HasPages() {
		res.Write("\n\n" + pager.Picker(p))
	}
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

func submenuURL(name string) string {
	// TODO: Submenu URL works because no other GET arguments are being used
	return "?submenu=" + name
}

func getCurrentSubmenu(rawURL string) string {
	_, rawQuery, found := strings.Cut(rawURL, "?")
	if !found {
		return ""
	}

	query, _ := url.ParseQuery(rawQuery)
	return query.Get("submenu")
}
