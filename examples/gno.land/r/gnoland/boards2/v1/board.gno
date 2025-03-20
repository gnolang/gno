package boards2

import (
	"net/url"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/gnoland/commondao"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/txlink"
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
	readonly  bool
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

// SetReadonly updates board's read-only status.
func (board *Board) SetReadonly(readonly bool) {
	board.readonly = readonly
}

// IsReadonly checks if the board is a read-only board.
func (board *Board) IsReadonly() bool {
	return board.readonly
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
func (board *Board) Render(path string) string {
	var sb strings.Builder

	sb.WriteString("# " + board.GetName() + "  \n")
	sb.WriteString("_Board created by " + newUserLink(board.creator) + " ")
	sb.WriteString("on " + board.createdAt.Format(dateFormat))
	sb.WriteString(", #" + board.id.String() + "_")

	if board.IsReadonly() {
		sb.WriteString("  \n_**Starting new threads and commenting is disabled**_")
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

	board.threads.ReverseIterateByOffset(p.Offset(), p.PageSize(), func(_ string, v any) bool {
		thread := v.(*Post)
		if thread.isHidden {
			return false
		}

		sb.WriteString(thread.RenderSummary())
		sb.WriteString("\n")
		return false
	})

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
	return txlink.Call("RenameBoard", "name", board.name)
}

func (board *Board) GetFreezeFormURL() string {
	return txlink.Call("FreezeBoard", "boardID", board.id.String())
}

func (board *Board) GetUnfreezeFormURL() string {
	return txlink.Call("UnfreezeBoard", "boardID", board.id.String())
}

func (board *Board) GetFlaggingThresholdFormURL() string {
	return txlink.Call("SetFlaggingThreshold", "boardID", board.id.String())
}

func (board *Board) GetInviteMemberFormURL() string {
	return txlink.Call("InviteMember", "boardID", board.id.String())
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
	perms.AddUser(owner, RoleOwner)
	return perms
}
