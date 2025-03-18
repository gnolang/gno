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

	// MaxReplyLength defines the maximum length allowed for replies.
	MaxReplyLength = 300
)

var (
	reBoardName = regexp.MustCompile(`(?i)^[a-z]+[a-z0-9_\-]{2,50}$`)

	// Minimalistic Markdown line prefix checks that if allowed would
	// break the current UI when submitting a reply. It denies replies
	// with headings, blockquotes or horizontal lines.
	reDeniedReplyLinePrefixes = regexp.MustCompile(`(?m)^\s*(#|---|>)+`)
)

// SetPermissions sets a permissions implementation for boards2 realm or a board.
func SetPermissions(bid BoardID, p Permissions) {
	assertRealmIsNotLocked()

	if p == nil {
		panic("permissions is required")
	}

	if bid != 0 {
		assertBoardExists(bid)
	}

	caller := std.PreviousRealm().Address()
	args := Args{bid}
	gPerms.WithPermission(caller, PermissionPermissionsUpdate, args, func(Args) {
		// When board ID is zero it means that realm permissions are being updated
		if bid == 0 {
			gPerms = p

			std.Emit(
				"RealmPermissionsUpdated",
				"caller", caller.String(),
			)
			return
		}

		// Otherwise update the permissions of a single board
		board := mustGetBoard(bid)
		board.perms = p

		std.Emit(
			"BoardPermissionsUpdated",
			"caller", caller.String(),
			"boardID", bid.String(),
		)
	})
}

// SetRealmNotice sets a notice to be displayed globally by the realm.
// An empty message removes the realm notice.
func SetRealmNotice(message string) {
	message = strings.TrimSpace(message)
	caller := std.PreviousRealm().Address()
	gPerms.WithPermission(caller, PermissionRealmNotice, nil, func(Args) {
		gNotice = message

		std.Emit(
			"RealmNoticeChanged",
			"caller", caller.String(),
			"message", message,
		)
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
	assertRealmIsNotLocked()

	name = strings.TrimSpace(name)
	assertIsValidBoardName(name)
	assertBoardNameNotExists(name)

	caller := std.PreviousRealm().Address()
	id := incGetBoardID()
	args := Args{name, id}
	gPerms.WithPermission(caller, PermissionBoardCreate, args, func(Args) {
		assertBoardNameNotExists(name)

		perms := createDefaultBoardPermissions(caller)
		board := newBoard(id, name, caller, perms)
		gBoardsByID.Set(id.Key(), board)
		gBoardsByName.Set(name, board)

		std.Emit(
			"BoardCreated",
			"caller", caller.String(),
			"boardID", id.String(),
			"name", name,
		)
	})
	return id
}

// RenameBoard changes the name of an existing board.
//
// A history of previous board names is kept when boards are renamed.
// Because of that boards are also accesible using previous name(s).
func RenameBoard(name, newName string) {
	assertRealmIsNotLocked()

	newName = strings.TrimSpace(newName)
	assertIsValidBoardName(newName)
	assertBoardNameNotExists(newName)

	board := mustGetBoardByName(name)
	assertBoardIsNotFrozen(board)

	bid := board.GetID()
	caller := std.PreviousRealm().Address()
	args := Args{bid, name, newName}
	board.perms.WithPermission(caller, PermissionBoardRename, args, func(Args) {
		assertBoardNameNotExists(newName)

		board := mustGetBoard(bid)
		board.aliases = append(board.aliases, board.name)
		board.name = newName

		// Index board for the new name keeping previous indexes for older names
		gBoardsByName.Set(newName, board)

		std.Emit(
			"BoardRenamed",
			"caller", caller.String(),
			"boardID", bid.String(),
			"name", name,
			"newName", newName,
		)
	})
}

// CreateThread creates a new thread within a board.
func CreateThread(boardID BoardID, title, body string) PostID {
	assertRealmIsNotLocked()

	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertHasBoardPermission(board, caller, PermissionThreadCreate)

	thread := board.AddThread(caller, title, body)

	std.Emit(
		"ThreadCreated",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", thread.GetPostID().String(),
		"title", title,
	)

	return thread.id
}

// CreateReply creates a new comment or reply within a thread.
//
// The value of `replyID` is only required when creating a reply of another reply.
func CreateReply(boardID BoardID, threadID, replyID PostID, body string) PostID {
	assertRealmIsNotLocked()

	body = strings.TrimSpace(body)
	assertReplyBodyIsValid(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertHasBoardPermission(board, caller, PermissionReplyCreate)

	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)
	assertThreadIsNotFrozen(thread)

	var reply *Post
	if replyID == 0 {
		// When the parent reply is the thread just add reply to thread
		reply = thread.AddReply(caller, body)
	} else {
		// Try to get parent reply and add a new child reply
		parent := mustGetReply(thread, replyID)
		if parent.IsHidden() || parent.IsReadOnly() {
			panic("replying to a hidden or frozen reply is not allowed")
		}

		reply = parent.AddReply(caller, body)
	}

	std.Emit(
		"ReplyCreate",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"replyID", reply.GetPostID().String(),
	)

	return reply.id
}

// CreateRepost reposts a thread into another board.
func CreateRepost(boardID BoardID, threadID PostID, title, body string, dstBoardID BoardID) PostID {
	assertRealmIsNotLocked()

	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	caller := std.PreviousRealm().Address()
	dst := mustGetBoard(dstBoardID)
	assertBoardIsNotFrozen(dst)
	assertHasBoardPermission(dst, caller, PermissionThreadRepost)

	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)

	if thread.IsRepost() {
		panic("reposting a thread that is a repost is not allowed")
	}

	repostId := dst.incGetPostID()
	body = strings.TrimSpace(body)
	repost := newPost(dst, repostId, caller, title, body, repostId, thread.GetPostID(), thread.GetBoard().GetID())
	thread.AddRepostTo(caller, repost, dst)

	std.Emit(
		"Repost",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"destinationBoardID", dstBoardID.String(),
		"title", title,
	)

	return repostId
}

// DeleteThread deletes a thread from a board.
//
// Threads can be deleted by the users who created them or otherwise by users with special permissions.
func DeleteThread(boardID BoardID, threadID PostID) {
	// Council members should always be able to delete
	caller := std.PreviousRealm().Address()
	isRealmOwner := gPerms.HasRole(caller, RoleOwner) // TODO: Add DeleteThread filetest cases for realm owners
	if !isRealmOwner {
		assertRealmIsNotLocked()
	}

	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)

	if !isRealmOwner {
		assertBoardIsNotFrozen(board)
		assertThreadIsNotFrozen(thread)

		if caller != thread.GetCreator() {
			assertHasBoardPermission(board, caller, PermissionThreadDelete)
		}
	}

	// Hard delete thread and all its replies
	board.DeleteThread(threadID)

	std.Emit(
		"ThreadDeleted",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
	)
}

// DeleteReply deletes a reply from a thread.
//
// Replies can be deleted by the users who created them or otherwise by users with special permissions.
// Soft deletion is used when the deleted reply contains sub replies, in which case the reply content
// is replaced by a text informing that reply has been deleted to avoid deleting sub-replies.
func DeleteReply(boardID BoardID, threadID, replyID PostID) {
	// Council members should always be able to delete
	caller := std.PreviousRealm().Address()
	isRealmOwner := gPerms.HasRole(caller, RoleOwner) // TODO: Add DeleteReply filetest cases for realm owners
	if !isRealmOwner {
		assertRealmIsNotLocked()
	}

	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)

	if !isRealmOwner {
		assertBoardIsNotFrozen(board)
		assertThreadIsNotFrozen(thread)
		assertReplyIsVisible(reply)
		assertReplyIsNotFrozen(reply)

		if caller != reply.GetCreator() {
			assertHasBoardPermission(board, caller, PermissionReplyDelete)
		}
	}

	// Soft delete reply by changing its body when it contains
	// sub-replies, otherwise hard delete it.
	if reply.HasReplies() {
		reply.Update(reply.GetTitle(), "This reply has been deleted")
	} else {
		thread.DeleteReply(replyID)
	}

	std.Emit(
		"ReplyDeleted",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"replyID", replyID.String(),
	)
}

// EditThread updates the title and body of thread.
//
// Threads can be updated by the users who created them or otherwise by users with special permissions.
func EditThread(boardID BoardID, threadID PostID, title, body string) {
	assertRealmIsNotLocked()

	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, threadID)
	assertThreadIsNotFrozen(thread)

	body = strings.TrimSpace(body)
	if !thread.IsRepost() {
		assertBodyIsNotEmpty(body)
	}

	caller := std.PreviousRealm().Address()
	if caller != thread.GetCreator() {
		assertHasBoardPermission(board, caller, PermissionThreadEdit)
	}

	thread.Update(title, body)

	std.Emit(
		"ThreadEdited",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"title", title,
	)
}

// EditReply updates the body of comment or reply.
//
// Replies can be updated only by the users who created them.
func EditReply(boardID BoardID, threadID, replyID PostID, body string) {
	assertRealmIsNotLocked()

	body = strings.TrimSpace(body)
	assertReplyBodyIsValid(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, threadID)
	assertThreadIsNotFrozen(thread)

	reply := mustGetReply(thread, replyID)
	assertReplyIsVisible(reply)
	assertReplyIsNotFrozen(reply)

	caller := std.PreviousRealm().Address()
	if caller != reply.GetCreator() {
		panic("only the reply creator is allowed to edit it")
	}

	reply.Update("", body)

	std.Emit(
		"ReplyEdited",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"replyID", replyID.String(),
		"body", body,
	)
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
	caller := std.PreviousRealm().Address()
	args := Args{user, role}
	perms.WithPermission(caller, PermissionMemberInvite, args, func(Args) {
		if err := perms.AddUser(user, role); err != nil {
			panic(err)
		}

		std.Emit(
			"MemberInvited",
			"caller", caller.String(),
			"boardID", boardID.String(),
			"member", user.String(),
			"role", string(role),
		)
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
	caller := std.PreviousRealm().Address()
	perms.WithPermission(caller, PermissionMemberRemove, Args{user}, func(Args) {
		if !perms.RemoveUser(user) {
			panic("member not found")
		}

		std.Emit(
			"MemberRemoved",
			"caller", caller.String(),
			"boardID", boardID.String(),
			"member", user.String(),
		)
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
	caller := std.PreviousRealm().Address()
	args := Args{boardID, member, role}
	perms.WithPermission(caller, PermissionRoleChange, args, func(Args) {
		if err := perms.SetUserRoles(member, role); err != nil {
			panic(err)
		}

		std.Emit(
			"RoleChanged",
			"caller", caller.String(),
			"boardID", boardID.String(),
			"member", member.String(),
			"newRole", string(role),
		)
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

func assertThreadIsNotFrozen(t *Post) {
	if t.IsReadOnly() {
		panic("thread is frozen")
	}
}

func assertReplyIsNotFrozen(r *Post) {
	if r.IsReadOnly() {
		panic("reply is frozen")
	}
}

func assertNameIsNotEmpty(name string) {
	if name == "" {
		panic("name is empty")
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

func assertThreadIsVisible(thread *Post) {
	if thread.IsHidden() {
		panic("thread is hidden")
	}
}

func assertReplyIsVisible(thread *Post) {
	if thread.IsHidden() {
		panic("reply is hidden")
	}
}

func assertReplyBodyIsValid(body string) {
	assertBodyIsNotEmpty(body)

	if len(body) > MaxReplyLength {
		n := strconv.Itoa(MaxReplyLength)
		panic("reply is too long, maximum allowed is " + n + " characters")
	}

	if reDeniedReplyLinePrefixes.MatchString(body) {
		panic("using Markdown headings, blockquotes or horizontal lines is not allowed in replies")
	}
}

func mustGetPermissions(bid BoardID) Permissions {
	if bid != 0 {
		board := mustGetBoard(bid)
		return board.perms
	}
	return gPerms
}
