package boards2

import (
	"regexp"
	"std"
	"strconv"
	"strings"
)

const (
	// MaxBoardNameLength defines the maximum length allowed for board names.
	MaxBoardNameLength = 50

	// MaxThreadTitleLength defines the maximum length allowed for thread titles.
	MaxThreadTitleLength = 100
)

var reBoardName = regexp.MustCompile(`(?i)^[a-z]+[a-z0-9_\-]{2,50}$`)

// SetPermissions sets a permissions implementation for boards2 realm or a board.
func SetPermissions(bid BoardID, p Permissions) {
	if p == nil {
		panic("permissions is required")
	}

	if bid != 0 {
		assertBoardExists(bid)
	}

	caller := std.OriginCaller()
	args := Args{bid}
	gPerms.WithPermission(caller, PermissionPermissionsUpdate, args, func(Args) {
		// When board ID is zero it means that realm permissions are being updated
		if bid == 0 {
			gPerms = p
			return
		}

		// Otherwise update the permissions of a single board
		board := mustGetBoard(bid)
		board.perms = p
	})
}

// GetBoardIDFromName searches a board by name and returns it's ID.
func GetBoardIDFromName(name string) (_ BoardID, found bool) {
	v, found := gBoardsByName.Get(name)
	if !found {
		return 0, false
	}
	return v.(*Board).id, true
}

// CreateBoard creates a new board.
func CreateBoard(name string) BoardID {
	name = strings.TrimSpace(name)
	assertIsValidBoardName(name)
	assertBoardNameNotExists(name)

	caller := std.OriginCaller()
	id := incGetBoardID()
	args := Args{name, id}
	gPerms.WithPermission(caller, PermissionBoardCreate, args, func(Args) {
		assertBoardNameNotExists(name)

		perms := createDefaultBoardPermissions(caller)
		board := newBoard(id, name, caller, perms)
		gBoardsByID.Set(id.Key(), board)
		gBoardsByName.Set(name, board)
	})
	return id
}

// RenameBoard changes the name of an existing board.
//
// A history of previous board names is kept when boards are renamed.
// Because of that boards are also accesible using previous name(s).
func RenameBoard(name, newName string) {
	newName = strings.TrimSpace(newName)
	assertIsValidBoardName(newName)
	assertBoardNameNotExists(newName)

	board := mustGetBoardByName(name)
	assertBoardIsNotFrozen(board)

	bid := board.GetID()
	caller := std.OriginCaller()
	args := Args{bid, name, newName}
	board.perms.WithPermission(caller, PermissionBoardRename, args, func(Args) {
		assertBoardNameNotExists(newName)

		board := mustGetBoard(bid)
		board.aliases = append(board.aliases, board.name)
		board.name = newName

		// Index board for the new name keeping previous indexes for older names
		gBoardsByName.Set(newName, board)
	})
}

// FreezeBoard freezes a board so no more threads and comments can be created or modified.
func FreezeBoard(boardID BoardID) {
	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.OriginCaller()
	args := Args{boardID}
	board.perms.WithPermission(caller, PermissionBoardFreeze, args, func(Args) {
		board := mustGetBoard(boardID)
		assertBoardIsNotFrozen(board)

		board.SetReadOnly(true)
	})
}

// IsBoardFrozen checks if a board has been frozen.
func IsBoardFrozen(boardID BoardID) bool {
	board := mustGetBoard(boardID)
	return board.IsReadOnly()
}

// SetFlaggingThreshold sets the number of flags required to hide a thread or comment.
//
// Threshold is only applicable within the board where it's setted.
func SetFlaggingThreshold(boardID BoardID, threshold int) {
	if threshold < 1 {
		panic("invalid flagging threshold")
	}

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.OriginCaller()
	args := Args{boardID, threshold}
	board.perms.WithPermission(caller, PermissionBoardFlaggingUpdate, args, func(Args) {
		assertBoardExists(boardID)

		gFlaggingThresholds.Set(boardID.String(), threshold)
	})
}

// GetFlaggingThreshold returns the number of flags required to hide a thread or comment within a board.
func GetFlaggingThreshold(boardID BoardID) int {
	assertBoardExists(boardID)
	return getFlaggingThreshold(boardID)
}

// FlagThread adds a new flag to a thread.
//
// Flagging requires special permissions and hides the thread when
// the number of flags reaches a pre-defined flagging threshold.
func FlagThread(boardID BoardID, threadID PostID, reason string) {
	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.OriginCaller()
	assertHasBoardPermission(board, caller, PermissionThreadFlag)

	t, ok := board.GetThread(threadID)
	if !ok {
		panic("post doesn't exist")
	}

	f := Flag{
		User:   caller,
		Reason: reason,
	}
	hide := flagItem(t, f, getFlaggingThreshold(boardID))
	if hide {
		t.SetVisible(false)
	}
}

// CreateThread creates a new thread within a board.
func CreateThread(boardID BoardID, title, body string) PostID {
	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.OriginCaller()
	assertHasBoardPermission(board, caller, PermissionThreadCreate)

	thread := board.AddThread(caller, title, body)
	return thread.id
}

// CreateReply creates a new comment or reply within a thread.
//
// The value of `replyID` is only required when creating a reply of another reply.
func CreateReply(boardID BoardID, threadID, replyID PostID, body string) PostID {
	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.OriginCaller()
	assertHasBoardPermission(board, caller, PermissionReplyCreate)

	thread := mustGetThread(board, threadID)
	assertThreadVisible(thread)

	var reply *Post
	if replyID == 0 {
		// When the parent reply is the thread just add reply to thread
		reply = thread.AddReply(caller, body)
	} else {
		// Try to get parent reply and add a new child reply
		parent := mustGetReply(thread, replyID)
		if parent.IsHidden() {
			panic("replying to a hidden reply is not allowed")
		}

		reply = parent.AddReply(caller, body)
	}
	return reply.id
}

// FlagReply adds a new flag to a comment or reply.
//
// Flagging requires special permissions and hides the comment or reply
// when the number of flags reaches a pre-defined flagging threshold.
func FlagReply(boardID BoardID, threadID, replyID PostID, reason string) {
	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.OriginCaller()
	assertHasBoardPermission(board, caller, PermissionThreadFlag)

	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)

	f := Flag{
		User:   caller,
		Reason: reason,
	}
	hide := flagItem(reply, f, getFlaggingThreshold(boardID))
	if hide {
		reply.SetVisible(false)
	}
}

// CreateRepost reposts a thread into another board.
func CreateRepost(boardID BoardID, threadID PostID, title, body string, dstBoardID BoardID) PostID {
	caller := std.OriginCaller()
	dst := mustGetBoard(dstBoardID)
	assertBoardIsNotFrozen(dst)
	assertHasBoardPermission(dst, caller, PermissionThreadRepost)

	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	assertThreadVisible(thread)

	if thread.IsRepost() {
		panic("reposting a thread that is a repost is not allowed")
	}

	repostId := dst.incGetPostID()
	repost := newPost(dst, repostId, caller, title, body, repostId, thread.GetPostID(), thread.GetBoard().GetID())
	thread.AddRepostTo(caller, repost, dst)
	return repostId
}

// DeleteThread deletes a thread from a board.
//
// Threads can be deleted by the users who created them or otherwise by users with special permissions.
func DeleteThread(boardID BoardID, threadID PostID) {
	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.OriginCaller()
	thread := mustGetThread(board, threadID)
	if caller != thread.GetCreator() {
		assertHasBoardPermission(board, caller, PermissionThreadDelete)
	}

	// TODO: Discuss how to deal with thread deletion (should we hide instead?)
	board.DeleteThread(threadID)
}

// DeleteReply deletes a reply from a thread.
//
// Replies can be deleted by the users who created them or otherwise by users with special permissions.
// Soft deletion is used when the deleted reply contains sub replies, in which case the reply content
// is replaced by a text informing that reply has been deleted to avoid deleting sub-replies.
func DeleteReply(boardID BoardID, threadID, replyID PostID) {
	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)
	caller := std.OriginCaller()
	if caller != reply.GetCreator() {
		assertHasBoardPermission(board, caller, PermissionReplyDelete)
	}

	// Soft delete reply by changing its body when it contains
	// sub-replies, otherwise hard delete it.
	if reply.HasReplies() {
		reply.Update(reply.GetTitle(), "This reply has been deleted")
	} else {
		thread.DeleteReply(replyID)
	}
}

// EditThread updates the title and body of thread.
//
// Threads can be updated by the users who created them or otherwise by users with special permissions.
func EditThread(boardID BoardID, threadID PostID, title, body string) {
	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, threadID)
	caller := std.OriginCaller()
	if caller != thread.GetCreator() {
		assertHasBoardPermission(board, caller, PermissionThreadEdit)
	}

	thread.Update(title, body)
}

// EditReply updates the body of comment or reply.
//
// Replies can be updated only by the users who created them.
func EditReply(boardID BoardID, threadID, replyID PostID, body string) {
	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)
	assertReplyVisible(reply)

	if std.OriginCaller() != reply.GetCreator() {
		panic("only the reply creator is allowed to edit it")
	}

	reply.Update("", body)
}

// InviteMember adds a member to the realm or to a boards.
//
// A role can optionally be specified to be assigned to the new member.
// Board ID is only required when inviting a member to a board.
func InviteMember(boardID BoardID, user std.Address, role Role) {
	if boardID != 0 {
		board := mustGetBoard(boardID)
		assertBoardIsNotFrozen(board)
	}

	perms := mustGetPermissions(boardID)
	caller := std.OriginCaller()
	args := Args{user, role}
	perms.WithPermission(caller, PermissionMemberInvite, args, func(Args) {
		if err := perms.AddUser(user, role); err != nil {
			panic(err)
		}
	})
}

// RemoveMember removes a member from the realm or a boards.
//
// Board ID is only required when removing a member from board.
func RemoveMember(boardID BoardID, user std.Address) {
	if boardID != 0 {
		board := mustGetBoard(boardID)
		assertBoardIsNotFrozen(board)
	}

	perms := mustGetPermissions(boardID)
	caller := std.OriginCaller()
	perms.WithPermission(caller, PermissionMemberRemove, Args{user}, func(Args) {
		if !perms.RemoveUser(user) {
			panic("member not found")
		}
	})
}

// IsMember checks if an user is a member of the realm or a board.
//
// Board ID is only required when checking if a user is a member of a board.
func IsMember(boardID BoardID, user std.Address) bool {
	if boardID != 0 {
		board := mustGetBoard(boardID)
		assertBoardIsNotFrozen(board)
	}

	perms := mustGetPermissions(boardID)
	return perms.HasUser(user)
}

// HasMemberRole checks if a realm or board member has a specific role assigned.
//
// Board ID is only required when checking a member of a board.
func HasMemberRole(boardID BoardID, member std.Address, role Role) bool {
	if boardID != 0 {
		board := mustGetBoard(boardID)
		assertBoardIsNotFrozen(board)
	}

	perms := mustGetPermissions(boardID)
	return perms.HasRole(member, role)
}

// ChangeMemberRole changes the role of a realm or board member.
//
// Board ID is only required when changing the role for a member of a board.
func ChangeMemberRole(boardID BoardID, member std.Address, role Role) {
	if boardID != 0 {
		board := mustGetBoard(boardID)
		assertBoardIsNotFrozen(board)
	}

	perms := mustGetPermissions(boardID)
	caller := std.OriginCaller()
	args := Args{boardID, member, role}
	perms.WithPermission(caller, PermissionRoleChange, args, func(Args) {
		if err := perms.SetUserRoles(member, role); err != nil {
			panic(err)
		}
	})
}

func assertHasBoardPermission(b *Board, user std.Address, p Permission) {
	if !b.perms.HasPermission(user, p) {
		panic("unauthorized")
	}
}

func assertBoardExists(id BoardID) {
	if _, found := getBoard(id); !found {
		panic("board not found: " + id.String())
	}
}

func assertBoardIsNotFrozen(b *Board) {
	if b.IsReadOnly() {
		panic("board is frozen")
	}
}

func assertIsValidBoardName(name string) {
	size := len(name)
	if size == 0 {
		panic("board name is empty")
	}

	if size > MaxBoardNameLength {
		n := strconv.Itoa(MaxBoardNameLength)
		panic("board name is too long, maximum allowed is " + n + " characters")
	}

	if !reBoardName.MatchString(name) {
		panic("board name contains invalid characters")
	}
}

func assertTitleIsValid(title string) {
	if title == "" {
		panic("title is empty")
	}

	if len(title) > MaxThreadTitleLength {
		n := strconv.Itoa(MaxThreadTitleLength)
		panic("thread title is too long, maximum allowed is " + n + " characters")
	}
}

func assertBodyIsNotEmpty(body string) {
	if body == "" {
		panic("body is empty")
	}
}

func assertBoardNameNotExists(name string) {
	if gBoardsByName.Has(name) {
		panic("board already exists")
	}
}

func assertThreadExists(b *Board, threadID PostID) {
	if _, found := b.GetThread(threadID); !found {
		panic("thread not found: " + threadID.String())
	}
}

func assertReplyExists(thread *Post, replyID PostID) {
	if _, found := thread.GetReply(replyID); !found {
		panic("reply not found: " + replyID.String())
	}
}

func assertThreadVisible(thread *Post) {
	if thread.IsHidden() {
		panic("thread is hidden")
	}
}

func assertReplyVisible(thread *Post) {
	if thread.IsHidden() {
		panic("reply is hidden")
	}
}

func mustGetPermissions(bid BoardID) Permissions {
	if bid != 0 {
		board := mustGetBoard(bid)
		return board.perms
	}
	return gPerms
}
