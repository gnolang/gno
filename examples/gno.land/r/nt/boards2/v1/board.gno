package boards2

import (
	"net/url"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/commondao"
)

type BoardID uint64

func (id BoardID) String() string {
	return strconv.Itoa(int(id))
}

func (id BoardID) Key() string {
	return padZero(uint64(id), 10)
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
	readOnly  bool
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
	return strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land") + ":" + url.PathEscape(board.name)
}

// GetURL returns relative board path relative.
//
// Note: returned result is not escaped. Use GetURL to get URL-encoded path.
func (board *Board) GetPath() string {
	return strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land") + ":" + board.name
}

func (board *Board) GetPermissions() Permissions {
	return board.perms
}

// SetReadOnly updates board's read-only status.
func (board *Board) SetReadOnly(readOnly bool) {
	board.readOnly = readOnly
}

// IsReadOnly checks if the board is a read-only board.
func (board *Board) IsReadOnly() bool {
	return board.readOnly
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
		panic("thread does not exist with id " + pid.String())
	}
}

// Render renders a board into Markdown.
//
// Pager is used for pagination if it's not nil.
func (board *Board) Render(p *PaginationOpts) string {
	var sb strings.Builder

	sb.WriteString("# Board: " + board.GetName() + "\n")
	sb.WriteString("> created by _" + board.creator.String() + "_ ")
	sb.WriteString("on " + board.createdAt.Format(dateFormat) + "\n\n")

	if board.threads.Size() == 0 {
		url := board.GetPostFormURL()
		sb.WriteString("### This board doesn't have any threads\n")
		sb.WriteString("Do you want to [create a new thread](" + url + ") in this board ?")
		return sb.String()
	}

	page := p.Iterate(&board.threads, func(_ string, v interface{}) bool {
		p := v.(*Post)
		if p.isHidden {
			return false
		}

		sb.WriteString("----------------------------------------\n")
		sb.WriteString(p.RenderSummary())
		sb.WriteString("\n")
		return false
	})

	if page != nil {
		sb.WriteString("\n---\n")
		sb.WriteString(page.Picker())
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

func (board *Board) GetRenameFormURL() string {
	return txlink.Call("RenameBoard", "name", board.name)
}

func (board *Board) GetFreezeFormURL() string {
	return txlink.Call("FreezeBoard", "boardID", board.id.String())
}

func (board *Board) GetFlaggingThresholdFormURL() string {
	return txlink.Call("SetFlaggingThreshold", "boardID", board.id.String())
}

func (board *Board) GetInviteMemberFormURL() string {
	return txlink.Call("InviteMember", "boardID", board.id.String())
}

func (board *Board) GetRemoveMemberFormURL() string {
	return txlink.Call("RemoveMember", "boardID", board.id.String())
}

func (board *Board) GetChangeMemberRoleFormURL() string {
	return txlink.Call("ChangeMemberRole", "boardID", board.id.String())
}

func (board *Board) GetPostFormURL() string {
	return txlink.Call("CreateThread", "boardID", board.id.String())
}

func (board *Board) GetMembersURL() string {
	return board.GetURL() + "/members"
}

func createDefaultBoardPermissions(owner std.Address) *DefaultPermissions {
	perms := NewDefaultPermissions(commondao.New())
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
		PermissionReplyCreate,
		PermissionReplyDelete,
		PermissionReplyFlag,
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
	perms.AddUser(owner, RoleOwner)
	return perms
}
