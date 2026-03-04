package boards2

import (
	"chain"
	"chain/runtime"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gno.land/p/gnoland/boards"
)

const (
	// MaxBoardNameLength defines the maximum length allowed for board names.
	MaxBoardNameLength = 50

	// MaxThreadTitleLength defines the maximum length allowed for thread titles.
	MaxThreadTitleLength = 100

	// MaxReplyLength defines the maximum length allowed for replies.
	MaxReplyLength = 1000
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
	caller := runtime.PreviousRealm().Address()
	args := boards.Args{content}
	gPerms.WithPermission(caller, PermissionRealmHelp, args, crossingFn(func() {
		gHelp = content
	}))
}

// SetPermissions sets a permissions implementation for boards2 realm or a board.
func SetPermissions(_ realm, boardID boards.ID, p boards.Permissions) {
	assertRealmIsNotLocked()
	assertBoardExists(boardID)

	if p == nil {
		panic("permissions is required")
	}

	caller := runtime.PreviousRealm().Address()
	args := boards.Args{boardID}
	gPerms.WithPermission(caller, PermissionPermissionsUpdate, args, crossingFn(func() {
		assertRealmIsNotLocked()

		// When board ID is zero it means that realm permissions are being updated
		if boardID == 0 {
			gPerms = p

			chain.Emit(
				"RealmPermissionsUpdated",
				"caller", caller.String(),
			)
			return
		}

		// Otherwise update the permissions of a single board
		board := mustGetBoard(boardID)
		board.Permissions = p

		chain.Emit(
			"BoardPermissionsUpdated",
			"caller", caller.String(),
			"boardID", board.ID.String(),
		)
	}))
}

// SetRealmNotice sets a notice to be displayed globally by the realm.
// An empty message removes the realm notice.
func SetRealmNotice(_ realm, message string) {
	message = strings.TrimSpace(message)
	caller := runtime.PreviousRealm().Address()
	args := boards.Args{message}
	gPerms.WithPermission(caller, PermissionRealmNotice, args, crossingFn(func() {
		gNotice = message

		chain.Emit(
			"RealmNoticeChanged",
			"caller", caller.String(),
			"message", message,
		)
	}))
}

// GetBoardIDFromName searches a board by name and returns its ID.
func GetBoardIDFromName(_ realm, name string) (_ boards.ID, found bool) {
	board, found := gBoards.GetByName(name)
	if !found {
		return 0, false
	}
	return board.ID, true
}

// CreateBoard creates a new board.
//
// Listed boards are included in the realm's list of boards.
// Open boards allow anyone to create threads and comment.
func CreateBoard(_ realm, name string, listed, open bool) boards.ID {
	assertRealmIsNotLocked()

	name = strings.TrimSpace(name)
	assertIsValidBoardName(name)
	assertBoardNameNotExists(name)

	caller := runtime.PreviousRealm().Address()
	id := gBoardsSequence.Next()
	board := boards.New(id)
	args := boards.Args{caller, name, board.ID, listed, open}
	gPerms.WithPermission(caller, PermissionBoardCreate, args, crossingFn(func() {
		assertRealmIsNotLocked()
		assertBoardNameNotExists(name)

		board.Name = name
		board.Creator = caller
		board.Meta = &Meta{
			HiddenThreads: boards.NewPostStorage(),
		}

		if open {
			board.Permissions = createOpenBoardPermissions(caller)
		} else {
			board.Permissions = createBasicBoardPermissions(caller)
		}

		if err := gBoards.Add(board); err != nil {
			panic(err)
		}

		// Listed boards are also indexed separately for easier iteration and pagination
		if listed {
			gListedBoardsByID.Set(board.ID.Key(), board)
		}

		chain.Emit(
			"BoardCreated",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"name", name,
		)
	}))
	return board.ID
}

// RenameBoard changes the name of an existing board.
//
// A history of previous board names is kept when boards are renamed.
// Because of that boards are also accessible using previous name(s).
func RenameBoard(_ realm, name, newName string) {
	assertRealmIsNotLocked()

	newName = strings.TrimSpace(newName)
	assertIsValidBoardName(newName)
	assertBoardNameNotExists(newName)

	board := mustGetBoardByName(name)
	assertBoardIsNotFrozen(board)

	caller := runtime.PreviousRealm().Address()
	args := boards.Args{caller, board.ID, name, newName}
	board.Permissions.WithPermission(caller, PermissionBoardRename, args, crossingFn(func() {
		assertRealmIsNotLocked()
		assertBoardNameNotExists(newName)

		board.Aliases = append(board.Aliases, board.Name)
		board.Name = newName

		// Index board for the new name keeping previous indexes for older names
		gBoards.Add(board)

		chain.Emit(
			"BoardRenamed",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"name", name,
			"newName", newName,
		)
	}))
}

// CreateThread creates a new thread within a board.
func CreateThread(_ realm, boardID boards.ID, title, body string) boards.ID {
	assertRealmIsNotLocked()

	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	caller := runtime.PreviousRealm().Address()
	assertUserIsNotBanned(boardID, caller)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	thread := boards.MustNewThread(board, caller, title, body)
	args := boards.Args{caller, board.ID, thread.ID, title, body}
	board.Permissions.WithPermission(caller, PermissionThreadCreate, args, crossingFn(func() {
		assertRealmIsNotLocked()
		assertUserIsNotBanned(board.ID, caller)

		thread.Replies = NewReplyStorage()

		if err := board.Threads.Add(thread); err != nil {
			panic(err)
		}

		chain.Emit(
			"ThreadCreated",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"threadID", thread.ID.String(),
			"title", title,
		)
	}))
	return thread.ID
}

// CreateReply creates a new comment or reply within a thread.
//
// The value of `replyID` is only required when creating a reply of another reply.
func CreateReply(_ realm, boardID, threadID, replyID boards.ID, body string) boards.ID {
	assertRealmIsNotLocked()

	body = strings.TrimSpace(body)
	assertReplyBodyIsValid(body)

	caller := runtime.PreviousRealm().Address()
	assertUserIsNotBanned(boardID, caller)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)
	assertThreadIsNotFrozen(thread)

	// By default consider that reply's parent is the thread.
	// Or when replyID is assigned use that reply as the parent.
	parent := thread
	if replyID > 0 {
		parent = mustGetReply(thread, replyID)
		if parent.Hidden || parent.Readonly {
			panic("replying to a hidden or frozen reply is not allowed")
		}
	}

	reply := boards.MustNewReply(parent, caller, body)
	args := boards.Args{caller, board.ID, thread.ID, parent.ID, reply.ID, body}
	board.Permissions.WithPermission(caller, PermissionReplyCreate, args, crossingFn(func() {
		assertRealmIsNotLocked()

		// Add reply to its parent
		if err := parent.Replies.Add(reply); err != nil {
			panic(err)
		}

		// When parent is not a thread also add reply to the thread.
		// The thread contains all replies and sub-replies, while each
		// reply only contains direct sub-replies.
		if parent.ID != thread.ID {
			if err := thread.Replies.Add(reply); err != nil {
				panic(err)
			}
		}

		chain.Emit(
			"ReplyCreate",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"threadID", thread.ID.String(),
			"replyID", reply.ID.String(),
		)
	}))
	return reply.ID
}

// CreateRepost reposts a thread into another board.
func CreateRepost(_ realm, boardID, threadID, destinationBoardID boards.ID, title, body string) boards.ID {
	assertRealmIsNotLocked()

	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	caller := runtime.PreviousRealm().Address()
	assertUserIsNotBanned(destinationBoardID, caller)

	dst := mustGetBoard(destinationBoardID)
	assertBoardIsNotFrozen(dst)

	board := mustGetBoard(boardID)
	thread := mustGetThread(board, threadID)
	assertThreadIsVisible(thread)

	repost := boards.MustNewRepost(thread, dst, caller)
	args := boards.Args{caller, board.ID, thread.ID, dst.ID, repost.ID, title, body}
	dst.Permissions.WithPermission(caller, PermissionThreadRepost, args, crossingFn(func() {
		assertRealmIsNotLocked()

		repost.Title = title
		repost.Body = strings.TrimSpace(body)

		if err := dst.Threads.Add(repost); err != nil {
			panic(err)
		}

		if err := thread.Reposts.Add(repost); err != nil {
			panic(err)
		}

		chain.Emit(
			"Repost",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"threadID", thread.ID.String(),
			"destinationBoardID", dst.ID.String(),
			"repostID", repost.ID.String(),
			"title", title,
		)
	}))
	return repost.ID
}

// DeleteThread deletes a thread from a board.
//
// Threads can be deleted by the users who created them or otherwise by users with special permissions.
func DeleteThread(_ realm, boardID, threadID boards.ID) {
	assertRealmIsNotLocked()

	caller := runtime.PreviousRealm().Address()
	board := mustGetBoard(boardID)
	assertUserIsNotBanned(boardID, caller)

	isRealmOwner := gPerms.HasRole(caller, RoleOwner) // TODO: Add DeleteThread filetest cases for realm owners
	if !isRealmOwner {
		assertBoardIsNotFrozen(board)
	}

	thread := mustGetThread(board, threadID)
	deleteThread := func() {
		board.Threads.Remove(thread.ID)

		chain.Emit(
			"ThreadDeleted",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"threadID", thread.ID.String(),
		)
	}

	// Thread can be directly deleted by user that created it.
	// It can also be deleted by realm owners, to be able to delete inappropriate content.
	// TODO: Discuss and decide if realm owners should be able to delete threads.
	if isRealmOwner || caller == thread.Creator {
		deleteThread()
		return
	}

	args := boards.Args{caller, board.ID, thread.ID}
	board.Permissions.WithPermission(caller, PermissionThreadDelete, args, crossingFn(func() {
		assertRealmIsNotLocked()
		deleteThread()
	}))
}

// DeleteReply deletes a reply from a thread.
//
// Replies can be deleted by the users who created them or otherwise by users with special permissions.
// Soft deletion is used when the deleted reply contains sub replies, in which case the reply content
// is replaced by a text informing that reply has been deleted to avoid deleting sub-replies.
func DeleteReply(_ realm, boardID, threadID, replyID boards.ID) {
	assertRealmIsNotLocked()

	caller := runtime.PreviousRealm().Address()
	board := mustGetBoard(boardID)
	assertUserIsNotBanned(boardID, caller)

	thread := mustGetThread(board, threadID)
	reply := mustGetReply(thread, replyID)
	isRealmOwner := gPerms.HasRole(caller, RoleOwner) // TODO: Add DeleteReply filetest cases for realm owners
	if !isRealmOwner {
		assertBoardIsNotFrozen(board)
		assertThreadIsNotFrozen(thread)
		assertReplyIsVisible(reply)
	}

	deleteReply := func() {
		// Soft delete reply by changing its body when it contains
		// sub-replies, otherwise hard delete it.
		if reply.Replies.Size() > 0 {
			reply.Body = "*This reply has been deleted*"
			reply.UpdatedAt = time.Now()
		} else {
			// Remove reply from the thread
			reply, removed := thread.Replies.Remove(replyID)
			if !removed {
				panic("reply not found")
			}

			// Remove reply from reply's parent
			if reply.ParentID != thread.ID {
				parent, found := thread.Replies.Get(reply.ParentID)
				if found {
					parent.Replies.Remove(replyID)
				}
			}
		}

		chain.Emit(
			"ReplyDeleted",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"threadID", thread.ID.String(),
			"replyID", reply.ID.String(),
		)
	}

	// Reply can be directly deleted by user that created it.
	// It can also be deleted by realm owners, to be able to delete inappropriate content.
	// TODO: Discuss and decide if realm owners should be able to delete replies.
	if isRealmOwner || caller == reply.Creator {
		deleteReply()
		return
	}

	args := boards.Args{caller, board.ID, thread.ID, reply.ID}
	board.Permissions.WithPermission(caller, PermissionReplyDelete, args, crossingFn(func() {
		assertRealmIsNotLocked()
		deleteReply()
	}))
}

// EditThread updates the title and body of a thread.
//
// Threads can be updated by the users who created them or otherwise by users with special permissions.
func EditThread(_ realm, boardID, threadID boards.ID, title, body string) {
	assertRealmIsNotLocked()

	title = strings.TrimSpace(title)
	assertTitleIsValid(title)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := runtime.PreviousRealm().Address()
	assertUserIsNotBanned(boardID, caller)

	thread := mustGetThread(board, threadID)
	assertThreadIsNotFrozen(thread)

	body = strings.TrimSpace(body)
	if !boards.IsRepost(thread) {
		assertBodyIsNotEmpty(body)
	}

	editThread := func() {
		thread.Title = title
		thread.Body = body
		thread.UpdatedAt = time.Now()

		chain.Emit(
			"ThreadEdited",
			"caller", caller.String(),
			"boardID", board.ID.String(),
			"threadID", thread.ID.String(),
			"title", title,
		)
	}

	if caller == thread.Creator {
		editThread()
		return
	}

	args := boards.Args{caller, board.ID, thread.ID, title, body}
	board.Permissions.WithPermission(caller, PermissionThreadEdit, args, crossingFn(func() {
		assertRealmIsNotLocked()
		editThread()
	}))
}

// EditReply updates the body of a comment or reply.
//
// Replies can be updated only by the users who created them.
func EditReply(_ realm, boardID, threadID, replyID boards.ID, body string) {
	assertRealmIsNotLocked()

	body = strings.TrimSpace(body)
	assertReplyBodyIsValid(body)

	board := mustGetBoard(boardID)
	assertBoardIsNotFrozen(board)

	caller := runtime.PreviousRealm().Address()
	assertUserIsNotBanned(boardID, caller)

	thread := mustGetThread(board, threadID)
	assertThreadIsNotFrozen(thread)

	reply := mustGetReply(thread, replyID)
	assertReplyIsVisible(reply)

	if caller != reply.Creator {
		panic("only the reply creator is allowed to edit it")
	}

	reply.Body = body
	reply.UpdatedAt = time.Now()

	chain.Emit(
		"ReplyEdited",
		"caller", caller.String(),
		"boardID", board.ID.String(),
		"threadID", thread.ID.String(),
		"replyID", reply.ID.String(),
		"body", body,
	)
}

// RemoveMember removes a member from the realm or a board.
//
// Board ID is only required when removing a member from board.
func RemoveMember(_ realm, boardID boards.ID, member address) {
	assertMembersUpdateIsEnabled(boardID)
	assertMemberAddressIsValid(member)

	perms := mustGetPermissions(boardID)
	origin := runtime.OriginCaller()
	caller := runtime.PreviousRealm().Address()
	removeMember := func() {
		if !perms.RemoveUser(member) {
			panic("member not found")
		}

		chain.Emit(
			"MemberRemoved",
			"caller", caller.String(),
			"origin", origin.String(), // When origin and caller match it means self removal
			"boardID", boardID.String(),
			"member", member.String(),
		)
	}

	// Members can remove themselves without permission
	if origin == member {
		removeMember()
		return
	}

	args := boards.Args{boardID, member}
	perms.WithPermission(caller, PermissionMemberRemove, args, crossingFn(func() {
		assertMembersUpdateIsEnabled(boardID)
		removeMember()
	}))
}

// IsMember checks if a user is a member of the realm or a board.
//
// Board ID is only required when checking if a user is a member of a board.
func IsMember(boardID boards.ID, user address) bool {
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
func HasMemberRole(boardID boards.ID, member address, role boards.Role) bool {
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
func ChangeMemberRole(_ realm, boardID boards.ID, member address, role boards.Role) {
	assertMemberAddressIsValid(member)
	assertMembersUpdateIsEnabled(boardID)

	if role == "" {
		role = RoleGuest
	}

	perms := mustGetPermissions(boardID)
	caller := runtime.PreviousRealm().Address()
	args := boards.Args{caller, boardID, member, role}
	perms.WithPermission(caller, PermissionRoleChange, args, crossingFn(func() {
		assertMembersUpdateIsEnabled(boardID)

		perms.SetUserRoles(member, role)

		chain.Emit(
			"RoleChanged",
			"caller", caller.String(),
			"boardID", boardID.String(),
			"member", member.String(),
			"newRole", string(role),
		)
	}))
}

// IterateRealmMembers iterates boards realm members.
// The iteration is done only for realm members, board members are not iterated.
func IterateRealmMembers(offset int, fn boards.UsersIterFn) (halted bool) {
	count := gPerms.UsersCount() - offset
	return gPerms.IterateUsers(offset, count, fn)
}

// GetBoard returns a single board.
func GetBoard(boardID boards.ID) *boards.Board {
	board := mustGetBoard(boardID)
	if !board.Permissions.HasRole(runtime.OriginCaller(), RoleOwner) {
		panic("forbidden")
	}
	return board
}

// Wraps a function to cross back to Boards2 realm.
func crossingFn(fn func()) func() {
	return func() {
		func(realm) { fn() }(cross)
	}
}

func assertMemberAddressIsValid(member address) {
	if !member.IsValid() {
		panic("invalid member address: " + member.String())
	}
}

func assertUserAddressIsValid(user address) {
	if !user.IsValid() {
		panic("invalid user address: " + user.String())
	}
}

func assertHasPermission(perms boards.Permissions, user address, p boards.Permission) {
	if !perms.HasPermission(user, p) {
		panic("unauthorized")
	}
}

func assertBoardExists(id boards.ID) {
	if id == 0 { // ID zero is used to refer to the realm
		return
	}

	if _, found := gBoards.Get(id); !found {
		panic("board not found: " + id.String())
	}
}

func assertBoardIsNotFrozen(b *boards.Board) {
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
		panic("board name must start with a letter and have letters, numbers, \"-\" and \"_\"")
	}
}

func assertThreadIsNotFrozen(t *boards.Post) {
	if t.Readonly {
		panic("thread is frozen")
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
		panic("title is too long, maximum allowed is " + n + " characters")
	}
}

func assertBodyIsNotEmpty(body string) {
	if body == "" {
		panic("body is empty")
	}
}

func assertBoardNameNotExists(name string) {
	name = strings.ToLower(name)
	if _, found := gBoards.GetByName(name); found {
		panic("board already exists")
	}
}

func assertThreadExists(b *boards.Board, threadID boards.ID) {
	if _, found := getThread(b, threadID); !found {
		panic("thread not found: " + threadID.String())
	}
}

func assertReplyExists(thread *boards.Post, replyID boards.ID) {
	if _, found := thread.Replies.Get(replyID); !found {
		panic("reply not found: " + replyID.String())
	}
}

func assertThreadIsVisible(thread *boards.Post) {
	if thread.Hidden {
		panic("thread is hidden")
	}
}

func assertReplyIsVisible(thread *boards.Post) {
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

func assertMembersUpdateIsEnabled(boardID boards.ID) {
	if boardID != 0 {
		assertRealmIsNotLocked()
	} else {
		assertRealmMembersAreNotLocked()
	}
}
