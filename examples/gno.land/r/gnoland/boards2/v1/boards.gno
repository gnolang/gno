package boards2

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
)

var (
	gPerms        Permissions
	gRealmLink    txlink.Realm
	gNotice       string
	gHelp         string
	gLastBoardID  BoardID
	gBoardsByID   avl.Tree // string(id) -> *Board
	gBoardsByName avl.Tree // string(name) -> *Board
	gLocked       struct {
		realm        bool
		realmMembers bool
	}
)

func init() {
	// Save current realm path so it's available during render calls
	gRealmLink = txlink.Realm(std.CurrentRealm().PkgPath())

	// Initialize default realm permissions
	gPerms = createDefaultPermissions(
		"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq", // @devx
		"g1manfred47kzduec920z88wfr64ylksmdcedlf5", // @moul
	)
}

// incGetBoardID returns a new board ID.
func incGetBoardID() BoardID {
	gLastBoardID++
	return gLastBoardID
}

// getBoard returns a board for a specific ID.
func getBoard(id BoardID) (_ *Board, found bool) {
	v, exists := gBoardsByID.Get(id.Key())
	if !exists {
		return nil, false
	}
	return v.(*Board), true
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
