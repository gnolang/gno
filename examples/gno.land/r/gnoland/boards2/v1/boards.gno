package boards2

import (
	"chain/runtime"

	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl"
)

// TODO: Refactor globals in favor of a cleaner pattern
var (
	gPerms            Permissions
	gRealmLink        txlink.Realm
	gNotice           string
	gHelp             string
	gLastReservedID   BoardID
	gBoardsByID       avl.Tree // string(id) -> *Board
	gListedBoardsByID avl.Tree // string(id) -> *Board
	gBoardsByName     avl.Tree // string(name) -> *Board
	gInviteRequests   avl.Tree // string(board id) -> *avl.Tree(std.Address -> time.Time)
	gBannedUsers      avl.Tree // string(board id) -> *avl.Tree(std.Address -> time.Time)
	gLocked           struct {
		realm        bool
		realmMembers bool
	}
)

func init() {
	// Save current realm path so it's available during render calls
	gRealmLink = txlink.Realm(runtime.CurrentRealm().PkgPath())

	// Initialize default realm permissions
	gPerms = initRealmPermissions(
		"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq", // @devx
		"g1manfred47kzduec920z88wfr64ylksmdcedlf5", // @moul
	)
}

// initRealmPermissions returns the default realm permissions.
func initRealmPermissions(owners ...address) *BasicPermissions {
	perms := createBasicPermissions(owners...)
	perms.ValidateFunc(PermissionBoardCreate, validateBoardCreate)
	perms.ValidateFunc(PermissionMemberInvite, validateMemberInvite)
	perms.ValidateFunc(PermissionRoleChange, validateRoleChange)
	return perms
}

// reserveBoardID returns a new board ID which won't be used by another board.
// Reserving IDs is necessary because boards can be created asynchronically.
func reserveBoardID() BoardID {
	gLastReservedID++
	return gLastReservedID
}

// getBoard returns a board for a specific ID.
func getBoard(id BoardID) (_ *Board, found bool) {
	v, exists := gBoardsByID.Get(id.Key())
	if !exists {
		return nil, false
	}
	return v.(*Board), true
}

// getInvitesRequests returns invite requests for a board.
func getInviteRequests(boardID BoardID) (_ *avl.Tree, found bool) {
	v, exists := gInviteRequests.Get(boardID.Key())
	if !exists {
		return nil, false
	}
	return v.(*avl.Tree), true
}

// getBannedUsers returns banned users within a board.
func getBannedUsers(boardID BoardID) (_ *avl.Tree, found bool) {
	v, exists := gBannedUsers.Get(boardID.Key())
	if !exists {
		return nil, false
	}
	return v.(*avl.Tree), true
}

// mustGetBoardByName returns a board or panics when it's not found.
func mustGetBoardByName(name string) *Board {
	v, found := gBoardsByName.Get(name)
	if !found {
		panic("board does not exist with name: " + name)
	}
	return v.(*Board)
}

// mustGetBoard returns a board or panics when it's not found.
func mustGetBoard(id BoardID) *Board {
	board, found := getBoard(id)
	if !found {
		panic("board does not exist with ID: " + id.String())
	}
	return board
}

// mustGetThread returns a thread or panics when it's not found.
func mustGetThread(board *Board, threadID PostID) *Post {
	thread, found := board.GetThread(threadID)
	if !found {
		panic("thread does not exist with ID: " + threadID.String())
	}
	return thread
}

// mustGetReply returns a reply or panics when it's not found.
func mustGetReply(thread *Post, replyID PostID) *Post {
	reply, found := thread.GetReply(replyID)
	if !found {
		panic("reply does not exist with ID: " + replyID.String())
	}
	return reply
}

func mustGetPermissions(bid BoardID) Permissions {
	if bid != 0 {
		board := mustGetBoard(bid)
		return board.perms
	}
	return gPerms
}

func parseRealmPath(path string) *realmpath.Request {
	// Make sure request is using current realm path so paths can be parsed during Render
	r := realmpath.Parse(path)
	r.Realm = string(gRealmLink)
	return r
}
