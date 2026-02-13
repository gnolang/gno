package boards2

import (
	"chain"
	"chain/runtime"
	"strings"
	"time"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/avl"
)

// Constants for different banning periods.
const (
	BanDay  = uint(24)
	BanWeek = BanDay * 7
	BanYear = BanDay * 365
)

// Ban bans a user from a board for a period of time.
// Only invited guest members and external users can be banned.
// Banning board owners, admins and moderators is not allowed.
func Ban(_ realm, boardID boards.ID, user address, hours uint, reason string) {
	assertAddressIsValid(user)

	if hours == 0 {
		panic("ban period in hours is required")
	}

	reason = strings.TrimSpace(reason)
	if reason == "" {
		panic("ban reason is required")
	}

	board := mustGetBoard(boardID)
	caller := runtime.PreviousRealm().Address()
	until := time.Now().Add(time.Minute * 60 * time.Duration(hours))
	args := boards.Args{boardID, user, until, reason}
	board.Permissions.WithPermission(caller, PermissionUserBan, args, crossingFn(func() {
		// When banning invited members make sure they are guests, otherwise
		// disallow banning. Only guest or external users can be banned.
		if board.Permissions.HasUser(user) && !board.Permissions.HasRole(user, RoleGuest) {
			panic("owner, admin and moderator banning is not allowed")
		}

		banned, found := getBannedUsers(boardID)
		if !found {
			banned = avl.NewTree()
			gBannedUsers.Set(boardID.Key(), banned)
		}

		banned.Set(user.String(), until)

		chain.Emit(
			"UserBanned",
			"bannedBy", caller.String(),
			"boardID", board.ID.String(),
			"user", user.String(),
			"until", until.Format(time.RFC3339),
			"reason", reason,
		)
	}))
}

// Unban unbans a user from a board.
func Unban(_ realm, boardID boards.ID, user address, reason string) {
	assertAddressIsValid(user)

	board := mustGetBoard(boardID)
	caller := runtime.PreviousRealm().Address()
	args := boards.Args{boardID, user, reason}
	board.Permissions.WithPermission(caller, PermissionUserUnban, args, crossingFn(func() {
		banned, found := getBannedUsers(boardID)
		if !found || !banned.Has(user.String()) {
			panic("user is not banned")
		}

		banned.Remove(user.String())

		chain.Emit(
			"UserUnbanned",
			"bannedBy", caller.String(),
			"boardID", board.ID.String(),
			"user", user.String(),
			"reason", reason,
		)
	}))
}

// IsBanned checks if a user is banned from a board.
func IsBanned(boardID boards.ID, user address) bool {
	banned, found := getBannedUsers(boardID)
	return found && banned.Has(user.String())
}

func assertAddressIsValid(addr address) {
	if !addr.IsValid() {
		panic("invalid address: " + addr.String())
	}
}

func assertUserIsNotBanned(boardID boards.ID, user address) {
	banned, found := getBannedUsers(boardID)
	if !found {
		return
	}

	v, found := banned.Get(user.String())
	if !found {
		return
	}

	until := v.(time.Time)
	if time.Now().Before(until) {
		panic(user.String() + " is banned until " + until.Format(dateFormat))
	}
}
