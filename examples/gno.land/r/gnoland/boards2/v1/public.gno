package boards2

import (
	"regexp"
	"std"
	"strconv"
	"strings"
	"time"
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

// SetHelp sets or updates boards realm help content.
func SetHelp(_ realm, content string) {
	content = strings.TrimSpace(content)
	caller := std.PreviousRealm().Address()
	args := Args{content}
	gPerms.WithPermission(cross, caller, PermissionRealmHelp, args, func(realm, Args) {
		gHelp = content
	})
}

// SetPermissions sets a permissions implementation for boards2 realm or a board.
func SetPermissions(_ realm, bid BoardID, p Permissions) {
	assertRealmIsNotLocked()

	if p == nil {
		panic("permissions is required")
	}

	if bid != 0 {
		assertBoardExists(bid)
	}

	caller := std.PreviousRealm().Address()
	args := Args{bid}
	gPerms.WithPermission(cross, caller, PermissionPermissionsUpdate, args, func(realm, Args) {
		assertRealmIsNotLocked()

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
func SetRealmNotice(_ realm, message string) {
	caller := std.PreviousRealm().Address()
	assertHasPermission(gPerms, caller, PermissionThreadCreate)

	gNotice = strings.TrimSpace(message)

	std.Emit(
		"RealmNoticeChanged",
		"caller", caller.String(),
		"message", gNotice,
	)
}

// GetBoardIDFromName searches a board by name and returns it's ID.
func GetBoardIDFromName(_ realm, name string) (_ BoardID, found bool) {
	v, found := gBoardsByName.Get(name)
	if !found {
		return 0, false
	}
	return v.(*Board).ID, true
}

// CreateBoard creates a new board.
//
// Listed boards are included in the list of boards.
func CreateBoard(_ realm, name string, listed bool) BoardID {
	assertRealmIsNotLocked()

	name = strings.TrimSpace(name)
	assertIsValidBoardName(name)
	assertBoardNameNotExists(name)

	caller := std.PreviousRealm().Address()
	id := reserveBoardID()
	args := Args{caller, name, id, listed}
	gPerms.WithPermission(cross, caller, PermissionBoardCreate, args, func(realm, Args) {
		assertRealmIsNotLocked()
		assertBoardNameNotExists(name)

		perms := createBasicBoardPermissions(caller)
		board := newBoard(id, name, caller, perms)
		key := id.Key()
		gBoardsByID.Set(key, board)
		gBoardsByName.Set(name, board)

		// Listed boards are also indexed separately for easier iteration and pagination
		if listed {
			gListedBoardsByID.Set(key, board)
		}

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
func RenameBoard(_ realm, name, newName string) {
	assertRealmIsNotLocked()

	newName = strings.TrimSpace(newName)
	assertIsValidBoardName(newName)
	assertBoardNameNotExists(newName)

	board := mustGetBoardByName(name)
	assertBoardIsNotFrozen(board)

	bid := board.ID
	caller := std.PreviousRealm().Address()
	args := Args{caller, bid, name, newName}
	board.perms.WithPermission(cross, caller, PermissionBoardRename, args, func(realm, Args) {
		assertRealmIsNotLocked()
		assertBoardNameNotExists(newName)

		board := mustGetBoard(bid)
		board.Aliases = append(board.Aliases, board.Name)
		board.Name = newName

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
func CreateThread(_ realm, boardID BoardID, title, body string) PostID {
	assertRealmIsNotLocked()

	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	body = strings.TrimSpace(body)
	assertBodyIsNotEmpty(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertUserIsNotBanned(board.ID, caller)
	assertHasPermission(board.perms, caller, PermissionThreadCreate)

	thread := board.AddThread(caller, title, body)

	std.Emit(
		"ThreadCreated",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", thread.ID.String(),
		"title", title,
	)

	return thread.ID
}

// CreateReply creates a new comment or reply within a thread.
//
// The value of `replyID` is only required when creating a reply of another reply.
func CreateReply(_ realm, boardID BoardID, threadID, replyID PostID, body string) PostID {
	assertRealmIsNotLocked()

	body = strings.TrimSpace(body)
	assertReplyBodyIsValid(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertHasPermission(board.perms, caller, PermissionReplyCreate)
	assertUserIsNotBanned(boardID, caller)

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
		if parent.Hidden || parent.Readonly {
			panic("replying to a hidden or frozen reply is not allowed")
		}

		reply = parent.AddReply(caller, body)
	}

	std.Emit(
		"ReplyCreate",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"replyID", reply.ID.String(),
	)

	return reply.ID
}

// CreateRepost reposts a thread into another board.
func CreateRepost(_ realm, boardID BoardID, threadID PostID, title, body string, destinationBoardID BoardID) PostID {
	assertRealmIsNotLocked()

	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	caller := std.PreviousRealm().Address()
	assertUserIsNotBanned(destinationBoardID, caller)

	dst := mustGetBoard(destinationBoardID)
	assertBoardIsNotFrozen(dst)
	assertHasPermission(dst.perms, caller, PermissionThreadRepost)

	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)

	if thread.IsRepost() {
		panic("reposting a thread that is a repost is not allowed")
	}

	body = strings.TrimSpace(body)
	repost := thread.Repost(caller, dst, title, body)

	std.Emit(
		"Repost",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"destinationBoardID", destinationBoardID.String(),
		"repostID", repost.ID.String(),
		"title", title,
	)

	return repost.ID
}

// DeleteThread deletes a thread from a board.
//
// Threads can be deleted by the users who created them or otherwise by users with special permissions.
func DeleteThread(_ realm, boardID BoardID, threadID PostID) {
	// Council members should always be able to delete
	caller := std.PreviousRealm().Address()
	isRealmOwner := gPerms.HasRole(caller, RoleOwner) // TODO: Add DeleteThread filetest cases for realm owners
	if !isRealmOwner {
		assertRealmIsNotLocked()
	}

	board := mustGetBoard(boardID)
	assertUserIsNotBanned(boardID, caller)

	thread := mustGetThread(board, threadID)

	if !isRealmOwner {
		assertBoardIsNotFrozen(board)
		assertThreadIsNotFrozen(thread)

		if caller != thread.Creator {
			assertHasPermission(board.perms, caller, PermissionThreadDelete)
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
func DeleteReply(_ realm, boardID BoardID, threadID, replyID PostID) {
	// Council members should always be able to delete
	caller := std.PreviousRealm().Address()
	isRealmOwner := gPerms.HasRole(caller, RoleOwner) // TODO: Add DeleteReply filetest cases for realm owners
	if !isRealmOwner {
		assertRealmIsNotLocked()
	}

	board := mustGetBoard(boardID)
	assertUserIsNotBanned(boardID, caller)

	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)

	if !isRealmOwner {
		assertBoardIsNotFrozen(board)
		assertThreadIsNotFrozen(thread)
		assertReplyIsVisible(reply)
		assertReplyIsNotFrozen(reply)

		if caller != reply.Creator {
			assertHasPermission(board.perms, caller, PermissionReplyDelete)
		}
	}

	// Soft delete reply by changing its body when it contains
	// sub-replies, otherwise hard delete it.
	if reply.HasReplies() {
		reply.Body = "This reply has been deleted"
		reply.UpdatedAt = time.Now()
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
func EditThread(_ realm, boardID BoardID, threadID PostID, title, body string) {
	assertRealmIsNotLocked()

	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertUserIsNotBanned(boardID, caller)

	thread := mustGetThread(board, threadID)
	assertThreadIsNotFrozen(thread)

	body = strings.TrimSpace(body)
	if !thread.IsRepost() {
		assertBodyIsNotEmpty(body)
	}

	if caller != thread.Creator {
		assertHasPermission(board.perms, caller, PermissionThreadEdit)
	}

	thread.Title = title
	thread.Body = body
	thread.UpdatedAt = time.Now()

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
func EditReply(_ realm, boardID BoardID, threadID, replyID PostID, body string) {
	assertRealmIsNotLocked()

	body = strings.TrimSpace(body)
	assertReplyBodyIsValid(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := std.PreviousRealm().Address()
	assertUserIsNotBanned(boardID, caller)

	thread := mustGetThread(board, threadID)
	assertThreadIsNotFrozen(thread)

	reply := mustGetReply(thread, replyID)
	assertReplyIsVisible(reply)
	assertReplyIsNotFrozen(reply)

	if caller != reply.Creator {
		panic("only the reply creator is allowed to edit it")
	}

	reply.Body = body
	reply.UpdatedAt = time.Now()

	std.Emit(
		"ReplyEdited",
		"caller", caller.String(),
		"boardID", boardID.String(),
		"threadID", threadID.String(),
		"replyID", replyID.String(),
		"body", body,
	)
}

// RemoveMember removes a member from the realm or a boards.
//
// Board ID is only required when removing a member from board.
func RemoveMember(_ realm, boardID BoardID, member std.Address) {
	assertMembersUpdateIsEnabled(boardID)
	assertMemberAddressIsValid(member)

	perms := mustGetPermissions(boardID)
	caller := std.PreviousRealm().Address()
	perms.WithPermission(cross, caller, PermissionMemberRemove, Args{member}, func(realm, Args) {
		assertMembersUpdateIsEnabled(boardID)

		if !perms.RemoveUser(cross, member) {
			panic("member not found")
		}

		std.Emit(
			"MemberRemoved",
			"caller", caller.String(),
			"boardID", boardID.String(),
			"member", member.String(),
		)
	})
}

// IsMember checks if an user is a member of the realm or a board.
//
// Board ID is only required when checking if a user is a member of a board.
func IsMember(boardID BoardID, user std.Address) bool {
	assertUserAddressIsValid(user)

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
	assertMemberAddressIsValid(member)

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
func ChangeMemberRole(_ realm, boardID BoardID, member std.Address, role Role) {
	assertMemberAddressIsValid(member)
	assertMembersUpdateIsEnabled(boardID)

	perms := mustGetPermissions(boardID)
	caller := std.PreviousRealm().Address()
	args := Args{caller, boardID, member, role}
	perms.WithPermission(cross, caller, PermissionRoleChange, args, func(realm, Args) {
		assertMembersUpdateIsEnabled(boardID)

		if err := perms.SetUserRoles(cross, member, role); err != nil {
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

// IterateRealmMembers iterates boards realm members.
// The iteration is done only for realm members, board members are not iterated.
func IterateRealmMembers(offset int, fn UsersIterFn) (halted bool) {
	count := gPerms.UsersCount() - offset
	return gPerms.IterateUsers(offset, count, fn)
}

// GetBoard returns a single board.
func GetBoard(boardID BoardID) *Board {
	board := mustGetBoard(boardID)
	if !board.perms.HasRole(std.OriginCaller(), RoleOwner) {
		panic("forbidden")
	}
	return board
}

func assertMemberAddressIsValid(member std.Address) {
	if !member.IsValid() {
		panic("invalid member address")
	}
}

func assertUserAddressIsValid(user std.Address) {
	if !user.IsValid() {
		panic("invalid user address")
	}
}

func assertHasPermission(perms Permissions, user std.Address, p Permission) {
	if !perms.HasPermission(user, p) {
		panic("unauthorized")
	}
}

func assertBoardExists(id BoardID) {
	if _, found := getBoard(id); !found {
		panic("board not found: " + id.String())
	}
}

func assertBoardIsNotFrozen(b *Board) {
	if b.Readonly {
		panic("board is frozen")
	}
}

func assertIsValidBoardName(name string) {
	size := len(name)
	if size == 0 {
		panic("board name is empty")
	}

	if size < 3 {
		panic("board name is too short, minimum length is 3 characters")
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
	if t.Readonly {
		panic("thread is frozen")
	}
}

func assertReplyIsNotFrozen(r *Post) {
	if r.Readonly {
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
	if thread.Hidden {
		panic("thread is hidden")
	}
}

func assertReplyIsVisible(thread *Post) {
	if thread.Hidden {
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

func assertMembersUpdateIsEnabled(boardID BoardID) {
	if boardID != 0 {
		assertRealmIsNotLocked()
	} else {
		assertRealmMembersAreNotLocked()
	}
}
