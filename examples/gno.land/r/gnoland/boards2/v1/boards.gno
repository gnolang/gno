package boards2

import (
	"chain/runtime"
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl"

	"gno.land/r/gnoland/boards2/v1/permissions"
)

// TODO: Refactor globals in favor of a cleaner pattern
var (
	gRealmLink        txlink.Realm
	gNotice           string
	gHelp             string
	gListedBoardsByID avl.Tree // string(id) -> *boards.Board
	gInviteRequests   avl.Tree // string(board id) -> *avl.Tree(address -> time.Time)
	gBannedUsers      avl.Tree // string(board id) -> *avl.Tree(address -> time.Time)
	gLocked           struct {
		realm        bool
		realmMembers bool
	}
)

var (
	gBoards         = boards.NewStorage()
	gBoardsSequence = boards.NewIdentifierGenerator()
	gRealmPath      = strings.TrimPrefix(runtime.CurrentRealm().PkgPath(), "gno.land")
	gPerms          = initRealmPermissions(
		"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq", // @devx
		"g1manfred47kzduec920z88wfr64ylksmdcedlf5", // @moul
	)

	// TODO: Allow updating open account amount though a proposal (GovDAO, CommonDAO?)
	gOpenAccountAmount = int64(3_000_000_000) // ugnot required for open board actions
)

func init() {
	// Save current realm path so it's available during render calls
	gRealmLink = txlink.Realm(runtime.CurrentRealm().PkgPath())
}

// initRealmPermissions returns the default realm permissions.
func initRealmPermissions(owners ...address) boards.Permissions {
	perms := permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole(permissions.RoleOwner),
	)
	perms.AddRole(permissions.RoleAdmin, permissions.PermissionBoardCreate)
	for _, owner := range owners {
		perms.SetUserRoles(cross, owner, permissions.RoleOwner)
	}

	perms.ValidateFunc(permissions.PermissionBoardCreate, validateBasicBoardCreate)
	perms.ValidateFunc(permissions.PermissionMemberInvite, validateBasicMemberInvite)
	perms.ValidateFunc(permissions.PermissionRoleChange, validateBasicRoleChange)
	return perms
}

// getInvitesRequests returns invite requests for a board.
func getInviteRequests(boardID boards.ID) (_ *avl.Tree, found bool) {
	v, exists := gInviteRequests.Get(boardID.Key())
	if !exists {
		return nil, false
	}
	return v.(*avl.Tree), true
}

// getBannedUsers returns banned users within a board.
func getBannedUsers(boardID boards.ID) (_ *avl.Tree, found bool) {
	v, exists := gBannedUsers.Get(boardID.Key())
	if !exists {
		return nil, false
	}
	return v.(*avl.Tree), true
}

// mustGetBoardByName returns a board or panics when it's not found.
func mustGetBoardByName(name string) *boards.Board {
	board, found := gBoards.GetByName(name)
	if !found {
		panic("board does not exist with name: " + name)
	}
	return board
}

// mustGetBoard returns a board or panics when it's not found.
func mustGetBoard(id boards.ID) *boards.Board {
	board, found := gBoards.Get(id)
	if !found {
		panic("board does not exist with ID: " + id.String())
	}
	return board
}

// mustGetThread returns a thread or panics when it's not found.
func mustGetThread(board *boards.Board, threadID boards.ID) *boards.Post {
	thread, found := board.Threads.Get(threadID)
	if !found {
		panic("thread does not exist with ID: " + threadID.String())
	}
	return thread
}

// mustGetReply returns a reply or panics when it's not found.
func mustGetReply(thread *boards.Post, replyID boards.ID) *boards.Post {
	reply, found := thread.Replies.Get(replyID)
	if !found {
		panic("reply does not exist with ID: " + replyID.String())
	}
	return reply
}

func mustGetPermissions(bid boards.ID) boards.Permissions {
	if bid != 0 {
		board := mustGetBoard(bid)
		return board.Permissions
	}
	return gPerms
}

func parseRealmPath(path string) *realmpath.Request {
	// Make sure request is using current realm path so paths can be parsed during Render
	r := realmpath.Parse(path)
	r.Realm = string(gRealmLink)
	return r
}
