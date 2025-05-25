package boards2

import (
	"net/url"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/nt/commondao"
)

type BoardID uint64

func (id BoardID) String() string {
	return strconv.Itoa(int(id))
}

func (id BoardID) Key() string {
	return seqid.ID(id).String()
}

type Board struct {
	id        BoardID // only set for public boards.
	name      string
	aliases   []string
	creator   std.Address
	threads   avl.Tree // Post.id -> *Post
	postsCtr  uint64   // increments Post.id
	createdAt time.Time
	deleted   avl.Tree // TODO reserved for fast-delete.
	perms     Permissions
	readonly  bool
	private   bool
}

func newBoard(id BoardID, name string, creator std.Address, p Permissions) *Board {
	return &Board{
		id:        id,
		name:      name,
		creator:   creator,
		threads:   avl.Tree{},
		createdAt: time.Now(),
		deleted:   avl.Tree{},
		perms:     p,
	}
}

func (board *Board) GetID() BoardID {
	return board.id
}

// GetName returns the name of the board.
func (board *Board) GetName() string {
	return board.name
}

// GetURL returns the relative URL of the board.
func (board *Board) GetURL() string {
	return strings.TrimPrefix(string(gRealmLink), "gno.land") + ":" + url.PathEscape(board.name)
}

// GetURL returns relative board path relative.
//
// Note: returned result is not escaped. Use GetURL to get URL-encoded path.
func (board *Board) GetPath() string {
	return strings.TrimPrefix(string(gRealmLink), "gno.land") + ":" + board.name
}

func (board *Board) GetPermissions() Permissions {
	return board.perms
}

// SetReadonly updates board's read-only status.
func (board *Board) SetReadonly(readonly bool) {
	board.readonly = readonly
}

// IsReadonly checks if the board is a read-only board.
func (board *Board) IsReadonly() bool {
	return board.readonly
}

// SetPrivate updates board's private status.
func (board *Board) SetPrivate(private bool) {
	board.private = private
}

// IsPrivate checks if the board is a private board.
func (board *Board) IsPrivate() bool {
	return board.private
}

func (board *Board) GetThread(threadID PostID) (_ *Post, found bool) {
	v, found := board.threads.Get(threadID.Key())
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

func (board *Board) AddThread(creator std.Address, title string, body string) *Post {
	pid := board.incGetPostID()
	thread := newPost(board, pid, creator, title, body, pid, 0, 0)
	board.threads.Set(pid.Key(), thread)
	return thread
}

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

	sb.WriteString("# " + board.GetName() + "  \n")
	sb.WriteString("Board created by " + newUserLink(board.creator) + " ")
	sb.WriteString("on " + board.createdAt.Format(dateFormat))
	sb.WriteString(", #" + board.id.String())

	if board.IsReadonly() {
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

	if board.threads.Size() == 0 {
		sb.WriteString("### This board doesn't have any threads\n")
		if !board.IsReadonly() {
			url := board.GetPostFormURL()
			sb.WriteString("Do you want to [start a new conversation](" + url + ") in this board ?")
		}
		return sb.String()
	}

	p, err := pager.New(path, board.threads.Size(), pager.WithPageSize(pageSizeThreads))
	if err != nil {
		panic(err)
	}

	render := func(_ string, v any) bool {
		thread := v.(*Post)
		if thread.isHidden {
			return false
		}

		sb.WriteString(thread.RenderSummary())
		sb.WriteString("\n")
		return false
	}

	sb.WriteString("Sort by: ")
	r := parseRealmPath(path)
	if r.Query.Get("order") == "desc" {
		r.Query.Set("order", "asc")
		sb.WriteString(newLink("newest first", r.String()))
		sb.WriteString("\n\n")
		board.threads.ReverseIterateByOffset(p.Offset(), p.PageSize(), render)
	} else {
		r.Query.Set("order", "desc")
		sb.WriteString(newLink("oldest first", r.String()))
		sb.WriteString("\n\n")
		board.threads.IterateByOffset(p.Offset(), p.PageSize(), render)
	}

	if p.HasPages() {
		sb.WriteString("\n---\n")
		sb.WriteString(pager.Picker(p))
	}

	return sb.String()
}

func (board *Board) incGetPostID() PostID {
	board.postsCtr++
	return PostID(board.postsCtr)
}

func (board *Board) GetURLFromThreadID(threadID PostID) string {
	return board.GetURL() + "/" + threadID.String()
}

func (board *Board) GetURLFromReplyID(threadID, replyID PostID) string {
	return board.GetURL() + "/" + threadID.String() + "/" + replyID.String()
}

func (board *Board) GetRenameFormURL() string { // TODO: These type of functions shouldn't be part of Board type
	return gRealmLink.Call(
		"RenameBoard",
		"name", board.name,
		"newName", "",
	)
}

func (board *Board) GetFreezeFormURL() string {
	return gRealmLink.Call(
		"FreezeBoard",
		"boardID", board.id.String(),
	)
}

func (board *Board) GetUnfreezeFormURL() string {
	return gRealmLink.Call(
		"UnfreezeBoard",
		"boardID", board.id.String(),
		"threadID", "",
		"replyID", "",
	)
}

func (board *Board) GetFlaggingThresholdFormURL() string {
	return gRealmLink.Call(
		"SetFlaggingThreshold",
		"boardID", board.id.String(),
		"threshold", "",
	)
}

func (board *Board) GetInviteMemberFormURL() string {
	return gRealmLink.Call(
		"InviteMember",
		"boardID", board.id.String(),
		"user", "",
		"role", "",
	)
}

func (board *Board) GetPostFormURL() string {
	return gRealmLink.Call(
		"CreateThread",
		"boardID", board.id.String(),
		"title", "",
		"body", "",
	)
}

func (board *Board) GetMembersURL() string {
	return board.GetURL() + "/members"
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
	return perms
}
