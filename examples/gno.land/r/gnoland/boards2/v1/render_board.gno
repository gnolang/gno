package boards2

import (
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/jeronimoalbi/mdform"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/md"
	"gno.land/p/nt/mdalert"
	"gno.land/p/nt/mux"
	"gno.land/r/gnoland/boards2/v1/permissions"
)

func renderBoard(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write(md.H3("The board you are looking for does not exist"))
		res.Write("Do you want to " + md.Link("create a new board", createBoardURI) + " ?")
		return
	}

	creatorLink := userLink(board.Creator)
	date := board.CreatedAt.Format(dateFormat)

	res.Write(md.H1("Boards › " + board.Name))
	if board.Readonly {
		res.Write(
			mdalert.Warning("Info", "Creating new threads and commenting are disabled within this board") + "\n",
		)
	}

	res.Write("Created by " + creatorLink + " on " + date + ", #" + board.ID.String())
	res.Write("  \n" + renderBoardMenu(board, req))
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

func renderBoardMenu(board *boards.Board, req *mux.Request) string {
	var (
		b               strings.Builder
		boardMembersURL = makeBoardURI(board) + "/members"
	)

	if board.Readonly {
		b.WriteString(md.Link("List Members", boardMembersURL))
		b.WriteString(" • ")
		b.WriteString(md.Link("Unfreeze Board", makeUnfreezeBoardURI(board)))
		b.WriteString("\n")
	} else {
		b.WriteString(md.Link("Create Thread", makeCreateThreadURI(board)))
		b.WriteString(" • ")
		b.WriteString(md.Link("Request Invite", makeRequestInviteURI(board)))
		b.WriteString(" • ")

		menu := getCurrentMenu(req.RawPath)
		if menu == menuManageBoard {
			b.WriteString(md.Bold("Manage Board"))
		} else {
			b.WriteString(md.Link("Manage Board", menuURL(menuManageBoard)))
		}

		b.WriteString("  \n")

		if menu == menuManageBoard {
			b.WriteString("↳")
			b.WriteString(md.Link("Invite Member", makeInviteMemberURI(board)))
			b.WriteString(" • ")
			b.WriteString(md.Link("List Invite Requests", makeBoardURI(board)+"/invites"))
			b.WriteString(" • ")
			b.WriteString(md.Link("List Members", boardMembersURL))
			b.WriteString(" • ")
			b.WriteString(md.Link("List Banned Users", makeBoardURI(board)+"/banned-users"))
			b.WriteString(" • ")
			b.WriteString(md.Link("Freeze Board", makeFreezeBoardURI(board)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	return b.String()
}

func renderInviteMember(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := gBoards.GetByName(name)
	if !found {
		res.Write("Board does not exist: " + name)
		return
	}

	form := mdform.New("exec", "InviteMember")
	form.Input(
		"boardID",
		"placeholder", "Board ID",
		"value", board.ID.String(),
		"readonly", "true",
	)
	form.Input(
		"user",
		"placeholder", "Address",
	)
	form.Select(
		"role",
		string(permissions.RoleOwner),
	)
	form.Select(
		"role",
		string(permissions.RoleAdmin),
	)
	form.Select(
		"role",
		string(permissions.RoleModerator),
	)
	form.Select(
		"role",
		string(permissions.RoleGuest),
		"selected", "true",
	)

	res.Write(md.H1(board.Name + ": Invite Member"))
	res.Write(
		md.Paragraph(
			"Both open and invite only boards can have multiple members with different roles within a "+
				"board, where members can have a single role at a time.",
		) +
			md.Paragraph(
				"Boards are independent communities which could apply different permissions per role than "+
					"other boards, but generally Boards2 supports four roles, _owner_, _admin_, _moderator_ "+
					"and _guest_.",
			) +
			md.Paragraph(
				"Member will be added to "+md.Link(board.Name, makeBoardURI(board))+" board.",
			),
	)
	res.Write(form.String())
}
