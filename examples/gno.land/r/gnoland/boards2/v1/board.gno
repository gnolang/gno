package boards2

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao"
)

type (
	// PostIterFn defines a function type to iterate posts.
	PostIterFn func(*Post) bool

	// BoardID defines a type for board identifiers.
	BoardID uint64
)

// String returns the ID as a string.
func (id BoardID) String() string {
	return strconv.Itoa(int(id))
}

// Key returns the ID as a string which can be used to index by ID.
func (id BoardID) Key() string {
	return seqid.ID(id).String()
}

// Board defines a type for boards.
type Board struct {
	ID       BoardID
	Name     string
	Aliases  []string
	Creator  std.Address
	Readonly bool
	Private  bool

	perms     Permissions
	postsCtr  uint64   // Increments Post.ID
	threads   avl.Tree // Post.ID -> *Post
	createdAt time.Time
}

func newBoard(id BoardID, name string, creator std.Address, p Permissions) *Board {
	return &Board{
		ID:        id,
		Name:      name,
		Creator:   creator,
		perms:     p,
		threads:   avl.Tree{},
		createdAt: time.Now(),
	}
}

// CreatedAt returns the time when board was created.
func (board *Board) CreatedAt() time.Time {
	return board.createdAt
}

// MembersCount returns the total number of board members.
func (board *Board) MembersCount() int {
	return board.perms.UsersCount()
}

// IterateMembers iterates board members.
func (board *Board) IterateMembers(start, count int, fn func(std.Address, []Role)) {
	board.perms.IterateUsers(start, count, func(u User) bool {
		fn(u.Address, u.Roles)
		return false
	})
}

// ThreadsCount returns the total number of board threads.
func (board *Board) ThreadsCount() int {
	return board.threads.Size()
}

// IterateThreads iterates board threads.
func (board *Board) IterateThreads(start, count int, fn PostIterFn) bool {
	return board.threads.IterateByOffset(start, count, func(_ string, v any) bool {
		p := v.(*Post)
		return fn(p)
	})
}

// ReverseIterateThreads iterates board threads in reverse order.
func (board *Board) ReverseIterateThreads(start, count int, fn PostIterFn) bool {
	return board.threads.ReverseIterateByOffset(start, count, func(_ string, v any) bool {
		p := v.(*Post)
		return fn(p)
	})
}

// GetThread returns board thread.
func (board *Board) GetThread(threadID PostID) (_ *Post, found bool) {
	v, found := board.threads.Get(threadID.Key())
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

// AddThread adds a new thread to the board.
func (board *Board) AddThread(creator std.Address, title, body string) *Post {
	pid := board.generateNextPostID()
	thread := newPost(board, pid, pid, creator, title, body)
	board.threads.Set(pid.Key(), thread)
	return thread
}

// DeleteThread deletes a thread from the board.
// NOTE: this can be potentially very expensive for threads with many replies.
// TODO: implement optional fast-delete where thread is simply moved.
func (board *Board) DeleteThread(pid PostID) {
	_, removed := board.threads.Remove(pid.Key())
	if !removed {
		panic("thread does not exist with ID " + pid.String())
	}
}

// Render renders a board into Markdown.
func (board *Board) Render(path, menu string) string {
	var sb strings.Builder

	sb.WriteString("# " + board.Name + "  \n")
	sb.WriteString("Board created by " + newUserLink(board.Creator) + " ")
	sb.WriteString("on " + board.CreatedAt().Format(dateFormat))
	sb.WriteString(", #" + board.ID.String())

	if board.Readonly {
		sb.WriteString("  \n_**Starting new threads and commenting is disabled**_")
	}

	// XXX: Menu is rendered by the caller to deal with links and sub-menus
	// TODO: We should have the render logic separated from boards so avoid sending menu as argument
	if menu != "" {
		sb.WriteString("\n\n")
		sb.WriteString(menu)
		sb.WriteString("\n")
	}

	sb.WriteString("\n\n---\n")

	if board.ThreadsCount() == 0 {
		sb.WriteString("### This board doesn't have any threads\n")
		if !board.Readonly {
			url := makeCreateThreadURI(board)
			sb.WriteString("Do you want to [start a new conversation](" + url + ") in this board ?")
		}
		return sb.String()
	}

	p, err := pager.New(path, board.ThreadsCount(), pager.WithPageSize(pageSizeThreads))
	if err != nil {
		panic(err)
	}

	render := func(thread *Post) bool {
		if !thread.Hidden {
			sb.WriteString(thread.RenderSummary())
			sb.WriteString("\n")
		}
		return false
	}

	sb.WriteString("Sort by: ")
	r := parseRealmPath(path)
	if r.Query.Get("order") == "desc" {
		r.Query.Set("order", "asc")
		sb.WriteString(md.Link("newest first", r.String()))
		sb.WriteString("\n\n")
		board.ReverseIterateThreads(p.Offset(), p.PageSize(), render)
	} else {
		r.Query.Set("order", "desc")
		sb.WriteString(md.Link("oldest first", r.String()))
		sb.WriteString("\n\n")
		board.IterateThreads(p.Offset(), p.PageSize(), render)
	}

	if p.HasPages() {
		sb.WriteString("\n---\n")
		sb.WriteString(pager.Picker(p))
	}

	return sb.String()
}

func (board *Board) generateNextPostID() PostID {
	board.postsCtr++
	return PostID(board.postsCtr)
}

func createBasicBoardPermissions(owner std.Address) *BasicPermissions {
	dao := commondao.New(commondao.WithMember(owner))
	perms := NewBasicPermissions(dao)
	perms.SetSuperRole(RoleOwner)
	perms.AddRole(
		RoleAdmin,
		PermissionBoardRename,
		PermissionBoardFlaggingUpdate,
		PermissionMemberInvite,
		PermissionMemberRemove,
		PermissionThreadCreate,
		PermissionThreadEdit,
		PermissionThreadDelete,
		PermissionThreadRepost,
		PermissionThreadFlag,
		PermissionThreadFreeze,
		PermissionReplyCreate,
		PermissionReplyDelete,
		PermissionReplyFlag,
		PermissionReplyFreeze,
		PermissionRoleChange,
	)
	perms.AddRole(
		RoleModerator,
		PermissionThreadCreate,
		PermissionThreadEdit,
		PermissionThreadRepost,
		PermissionThreadFlag,
		PermissionReplyCreate,
		PermissionReplyFlag,
	)
	perms.AddRole(
		RoleGuest,
		PermissionThreadCreate,
		PermissionThreadRepost,
		PermissionReplyCreate,
	)
	perms.SetUserRoles(owner, RoleOwner)
	perms.ValidateFunc(PermissionBoardRename, validateBoardRename)
	perms.ValidateFunc(PermissionMemberInvite, validateMemberInvite)
	perms.ValidateFunc(PermissionRoleChange, validateRoleChange)
	return perms
}
