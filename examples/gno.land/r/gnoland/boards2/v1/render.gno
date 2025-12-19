package boards2

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"gno.land/p/gnoland/boards"
	"gno.land/p/jeronimoalbi/mdform"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/mux"
)

const (
	pageSizeDefault = 6
	pageSizeReplies = 10
)

const menuManageBoard = "manageBoard"

var (
	createBoardURI = gRealmPath + ":create-board"
	adminUsersURI  = gRealmPath + ":admin-users"
	helpURI        = gRealmPath + ":help"
)

func Render(path string) string {
	var (
		b      strings.Builder
		router = mux.NewRouter()
	)

	router.HandleFunc("", renderBoardsList)
	router.HandleFunc("help", renderHelp)
	router.HandleFunc("admin-users", renderMembers)
	router.HandleFunc("create-board", renderCreateBoard)
	router.HandleFunc("{board}", renderBoard)
	router.HandleFunc("{board}/members", renderMembers)
	router.HandleFunc("{board}/invites", renderInvites)
	router.HandleFunc("{board}/banned-users", renderBannedUsers)
	router.HandleFunc("{board}/create-thread", renderCreateThread)
	router.HandleFunc("{board}/invite-member", renderInviteMember)
	router.HandleFunc("{board}/{thread}", renderThread)
	router.HandleFunc("{board}/{thread}/flag", renderFlagPost)
	router.HandleFunc("{board}/{thread}/reply", renderReplyPost)
	router.HandleFunc("{board}/{thread}/edit", renderEditThread)
	router.HandleFunc("{board}/{thread}/repost", renderRepostThread)
	router.HandleFunc("{board}/{thread}/{reply}", renderReply)
	router.HandleFunc("{board}/{thread}/{reply}/flag", renderFlagPost)
	router.HandleFunc("{board}/{thread}/{reply}/reply", renderReplyPost)
	router.HandleFunc("{board}/{thread}/{reply}/edit", renderEditReply)

	router.NotFoundHandler = func(res *mux.ResponseWriter, _ *mux.Request) {
		res.Write(md.Blockquote("Path not found"))
	}

	// Render common realm header before resolving render path
	if gNotice != "" {
		b.WriteString(infoAlert("Notice", gNotice))
	}

	// Render view for current path
	b.WriteString(router.Render(path))

	return b.String()
}

func renderHelp(res *mux.ResponseWriter, _ *mux.Request) {
	res.Write(md.H1("Boards Help"))
	if gHelp != "" {
		res.Write(gHelp)
		return
	}

	link := gRealmLink.Call("SetHelp", "content", "")
	res.Write(md.H3("Help content has not been uploaded"))
	res.Write("Do you want to " + md.Link("upload boards help", link) + " ?")
}

func renderBoardsList(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(md.H1("Boards"))
	renderBoardListMenu(res, req)
	res.Write(md.HorizontalRule())

	if gListedBoardsByID.Size() == 0 {
		res.Write(md.H3("Currently there are no boards"))
		res.Write("Be the first to " + md.Link("create a new board", createBoardURI) + " !")
		return
	}

	p, err := pager.New(req.RawPath, gListedBoardsByID.Size(), pager.WithPageSize(pageSizeDefault))
	if err != nil {
		panic(err)
	}

	render := func(_ string, v any) bool {
		board := v.(*boards.Board)
		userLink := userLink(board.Creator)
		date := board.CreatedAt.Format(dateFormat)

		res.Write(md.Bold(md.Link(board.Name, makeBoardURI(board))) + "  \n")
		res.Write("Created by " + userLink + " on " + date + ", #" + board.ID.String() + "  \n")

		status := strconv.Itoa(board.Threads.Size()) + " threads"
		if board.Readonly {
			status += ", read-only"
		}

		res.Write(md.Bold(status) + "\n\n")
		return false
	}

	res.Write("Sort by: ")
	r := parseRealmPath(req.RawPath)
	if r.Query.Get("order") == "desc" {
		r.Query.Set("order", "asc")
		res.Write(md.Link("newest first", r.String()) + "\n\n")
		gListedBoardsByID.ReverseIterateByOffset(p.Offset(), p.PageSize(), render)
	} else {
		r.Query.Set("order", "desc")
		res.Write(md.Link("oldest first", r.String()) + "\n\n")
		gListedBoardsByID.IterateByOffset(p.Offset(), p.PageSize(), render)
	}

	if p.HasPages() {
		res.Write(md.HorizontalRule())
		res.Write(pager.Picker(p))
	}
}

func renderBoardListMenu(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(md.Link("Create Board", createBoardURI))
	res.Write(" • ")
	res.Write(md.Link("List Admin Users", adminUsersURI))
	res.Write(" • ")
	res.Write(md.Link("Help", helpURI))
	res.Write("\n\n")
}

func renderCreateBoard(res *mux.ResponseWriter, _ *mux.Request) {
	form := mdform.New("exec", "CreateBoard")
	form.Input(
		"name",
		"placeholder", "Board name",
	)
	form.Radio(
		"listed",
		"true",
		"checked", "true",
		"description", "Should board be publically listed?",
	)
	form.Radio(
		"listed",
		"false",
	)
	form.Radio(
		"open",
		"true",
		"description", "Should anyone be allowed to create threads and comments?",
	)
	form.Radio(
		"open",
		"false",
		"checked", "true",
	)

	res.Write(md.H1("Boards: Create Board"))
	res.Write(
		md.Paragraph(
			"Boards are by default listed by the realm but they can optionally " +
				"be created so they are only found by their URL.",
		),
	)
	res.Write(
		md.Paragraph(
			"They can also be created to be open so anyone is allowed to create " +
				"new threads and also to comment on any thread within the open board.",
		),
	)
	res.Write(form.String())
}

func renderMembers(res *mux.ResponseWriter, req *mux.Request) {
	boardID := boards.ID(0)
	perms := gPerms
	name := req.GetVar("board")
	if name != "" {
		board, found := gBoards.GetByName(name)
		if !found {
			res.Write(md.H3("Board not found"))
			return
		}

		boardID = board.ID
		perms = board.Permissions

		res.Write(md.H1(board.Name + " Members"))
		res.Write(md.H3("These are the board members"))
	} else {
		res.Write(md.H1("Admin Users"))
		res.Write(md.H3("These are the admin users of the realm"))
	}

	// Create a pager with a small page size to reduce
	// the number of username lookups per page.
	p, err := pager.New(req.RawPath, perms.UsersCount(), pager.WithPageSize(pageSizeDefault))
	if err != nil {
		res.Write(err.Error())
		return
	}

	table := mdtable.Table{
		Headers: []string{"Member", "Role", "Actions"},
	}

	perms.IterateUsers(p.Offset(), p.PageSize(), func(u boards.User) bool {
		actions := []string{
			md.Link("remove", gRealmLink.Call(
				"RemoveMember",
				"boardID", boardID.String(),
				"member", u.Address.String(),
			)),
			md.Link("change role", gRealmLink.Call(
				"ChangeMemberRole",
				"boardID", boardID.String(),
				"member", u.Address.String(),
				"role", "",
			)),
		}

		table.Append([]string{
			userLink(u.Address),
			rolesToString(u.Roles),
			strings.Join(actions, " • "),
		})
		return false
	})
	res.Write(table.String())

	if p.HasPages() {
		res.Write("\n" + pager.Picker(p))
	}
}

func renderInvites(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write(md.H3("Board not found"))
		return
	}

	res.Write(md.H1(board.Name + " Invite Requests"))

	requests, found := getInviteRequests(board.ID)
	if !found || requests.Size() == 0 {
		res.Write(md.H3("Board has no invite requests"))
		return
	}

	p, err := pager.New(req.RawPath, requests.Size(), pager.WithPageSize(pageSizeDefault))
	if err != nil {
		res.Write(err.Error())
		return
	}

	table := mdtable.Table{
		Headers: []string{"User", "Request Date", "Actions"},
	}

	res.Write(md.H3("These users have requested to be invited to the board"))
	requests.ReverseIterateByOffset(p.Offset(), p.PageSize(), func(addr string, v any) bool {
		actions := []string{
			md.Link("accept", gRealmLink.Call(
				"AcceptInvite",
				"boardID", board.ID.String(),
				"user", addr,
			)),
			md.Link("revoke", gRealmLink.Call(
				"RevokeInvite",
				"boardID", board.ID.String(),
				"user", addr,
			)),
		}

		table.Append([]string{
			userLink(address(addr)),
			v.(time.Time).Format(dateFormat),
			strings.Join(actions, " • "),
		})
		return false
	})

	res.Write(table.String())

	if p.HasPages() {
		res.Write("\n" + pager.Picker(p))
	}
}

func renderBannedUsers(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write(md.H3("Board not found"))
		return
	}

	res.Write(md.H1(board.Name + " Banned Users"))

	banned, found := getBannedUsers(board.ID)
	if !found || banned.Size() == 0 {
		res.Write(md.H3("Board has no banned users"))
		return
	}

	p, err := pager.New(req.RawPath, banned.Size(), pager.WithPageSize(pageSizeDefault))
	if err != nil {
		res.Write(err.Error())
		return
	}

	table := mdtable.Table{
		Headers: []string{"User", "Banned Until", "Actions"},
	}

	res.Write(md.H3("These users have been banned from the board"))
	banned.ReverseIterateByOffset(p.Offset(), p.PageSize(), func(addr string, v any) bool {
		table.Append([]string{
			userLink(address(addr)),
			v.(time.Time).Format(dateFormat),
			md.Link("unban", gRealmLink.Call(
				"Unban",
				"boardID", board.ID.String(),
				"user", addr,
				"reason", "",
			)),
		})
		return false
	})

	res.Write(table.String())

	if p.HasPages() {
		res.Write("\n" + pager.Picker(p))
	}
}

func infoAlert(title, msg string) string {
	header := strings.TrimSpace("[!INFO] " + title)
	return md.Blockquote(header + "\n" + msg)
}

func rolesToString(roles []boards.Role) string {
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
