package boards2

import (
	"chain/runtime"
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl/v0"
)

var (
	// RealmLink contains Boards2 realm link.
	// It can be used to generate board TX links from other realms.
	RealmLink = txlink.Realm(runtime.CurrentRealm().PkgPath())

	// RequiredAccountAmount contains the required account amount for open board interactions.
	// The amount requirement is not applied to members that were invited to an open board.
	// Amount is defined as ugnot.
	RequiredAccountAmount = int64(3_000_000_000)

	// Notice contains an optional message that is displayed globally within the realm.
	Notice string

	// Help contains optional Markdown with Boards2 realm help.
	Help string
)

// TODO: Refactor globals in favor of a cleaner pattern
var (
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
	gPerms          = initRealmPermissions("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh") // GovDAO T1 multisig
)

// initRealmPermissions returns the default realm permissions.
func initRealmPermissions(owners ...address) boards.Permissions {
	perms := permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole(RoleOwner),
	)
	perms.AddRole(RoleAdmin, PermissionBoardCreate)
	for _, owner := range owners {
		perms.SetUserRoles(owner, RoleOwner)
	}

	perms.ValidateFunc(PermissionBoardCreate, validateBasicBoardCreate)
	perms.ValidateFunc(PermissionMemberInvite, validateBasicMemberInvite)
	perms.ValidateFunc(PermissionRoleChange, validateBasicRoleChange)
	return perms
}

// getInviteRequests returns invite requests for a board.
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

// getThread returns a board thread.
func getThread(board *boards.Board, threadID boards.ID) (*boards.Post, bool) {
	thread, found := board.Threads.Get(threadID)
	if !found {
		// When thread is not found search it within hidden threads
		meta := board.Meta.(*BoardMeta)
		thread, found = meta.HiddenThreads.Get(threadID)
	}
	return thread, found
}

// getReply returns a thread comment or reply.
func getReply(thread *boards.Post, replyID boards.ID) (*boards.Post, bool) {
	meta := thread.Meta.(*ThreadMeta)
	return meta.AllReplies.Get(replyID)
}

// mustGetThread returns a thread or panics when it's not found.
func mustGetThread(board *boards.Board, threadID boards.ID) *boards.Post {
	thread, found := getThread(board, threadID)
	if !found {
		panic("thread does not exist with ID: " + threadID.String())
	}
	return thread
}

// mustGetReply returns a reply or panics when it's not found.
func mustGetReply(thread *boards.Post, replyID boards.ID) *boards.Post {
	reply, found := getReply(thread, replyID)
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
	r.Realm = string(RealmLink)
	return r
}
